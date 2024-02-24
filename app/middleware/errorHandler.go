package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type stackError interface {
	StackTrace() errors.StackTrace
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	for _, err := range c.Errors {
		if serverError, found := serverErrorFromCause(err.Err); found {
			log.Warnf("Handler returned error, apperrors.ServerError found in error chain: %d %s",
				serverError.Code, serverError.Error())
			log.Infof("Error: %+v", err.Err)
			c.JSON(serverError.Code, gin.H{"error": serverError.Message})
			return
		}
	}

	if len(c.Errors) == 1 {
		log.WithError(c.Errors.Last().Err).Errorf(
			"Handler returned error, apperrors.ServerError not found in error chain, responding with 500")
		err := c.Errors.Last().Err
		for err != nil {
			log.Infof("Cause: %s", err.Error())
			if stackErr, ok := err.(stackError); ok {
				log.Infof("Error stacktrace: %+v", stackErr.StackTrace())
			}
			err = errors.Unwrap(err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
		return
	}
	log.WithField("errors", c.Errors.JSON()).Errorln(
		"Handler returned multiple errors, apperrors.ServerError not found in error chain, responding with 500")
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
}

func serverErrorFromCause(err error) (serverError *apperrors.ServerError, found bool) {
	for err != nil {
		if e, ok := err.(*apperrors.ServerError); ok {
			return e, true
		}
		err = errors.Unwrap(err)
	}
	return nil, false
}
