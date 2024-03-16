package server

import (
	"bufio"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/core"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"time"
)

func ServiceLogStreamEndpoint(r *gin.Engine, c *core.Core, rdb *redis.Client) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	r.GET("/api/v1/services/:id/logs", func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			log.Errorln("Token is not provided")
			_ = ctx.AbortWithError(http.StatusUnauthorized, apperrors.Unauthorized("Token is not provided"))
			return
		}

		username, err := rdb.Get(ctx, token).Result()
		if err != nil {
			log.WithError(err).Errorln("Invalid token provided")
			_ = ctx.AbortWithError(http.StatusForbidden, apperrors.ForbiddenWrap(err, "Invalid token provided"))
			return
		}
		rdb.Del(ctx, token)

		idStr := ctx.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, apperrors.BadRequest("Failed to parse service id"))
		}

		replicaStr := ctx.Query("replica")
		replica := 0
		if replicaStr != "" {
			replica, err = strconv.Atoi(replicaStr)
			if err != nil {
				_ = ctx.AbortWithError(http.StatusBadRequest, apperrors.BadRequest("Failed to parse replica index"))
			}
		}

		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.WithError(err).Errorln("Failed to upgrade to WebSocket")
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		logCtx, cancel := context.WithCancel(ctx)

		logs, err := c.Services.StreamServiceLogs(logCtx, id, replica, middleware.Authentication{Username: username})
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		r := bufio.NewReader(logs)

		conn.SetCloseHandler(func(code int, text string) error {
			log.Debugln("WebSocket connection closed")
			cancel()
			return nil
		})

		startMonitorConn(conn, cancel)

		for {
			line, isPrefix, err := r.ReadLine()
			if err != nil {
				log.WithError(err).Errorln("Failed to read service logs")
				_ = conn.Close()
				cancel()
				return
			}
			if !isPrefix {
				line = append(line, '\n')
			}
			_ = conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
			err = conn.WriteMessage(websocket.TextMessage, line)
			if err != nil {
				log.WithError(err).Errorln("Failed to write service logs to WebSocket")
				_ = conn.Close()
				cancel()
				return
			}
		}
	})
}

func startMonitorConn(conn *websocket.Conn, cancel context.CancelFunc) {
	go func() {
		for {
			err := conn.SetReadDeadline(time.Now().Add(20 * time.Second))
			if err != nil {
				log.WithError(err).Errorln("WebSocket ping read timed out")
				_ = conn.Close()
				cancel()
				return
			}
			msgType, reader, err := conn.NextReader()
			if err != nil {
				log.WithError(err).Errorln("WebSocket ping read failed")
				_ = conn.Close()
				cancel()
				return
			}
			if msgType != websocket.TextMessage {
				log.Errorln("WebSocket ping unknown message")
				_ = conn.Close()
				cancel()
				return
			}
			data, err := io.ReadAll(reader)
			if err != nil {
				log.WithError(err).Errorln("WebSocket ping read failed")
				_ = conn.Close()
				cancel()
				return
			}
			if string(data) != "ping" {
				log.Errorf("WebSocket ping unknown message: %s", string(data))
				_ = conn.Close()
				cancel()
				return
			}
			log.Debugln("Ping received")
		}
	}()
}
