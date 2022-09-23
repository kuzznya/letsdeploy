package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/appErrors"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type stackError interface {
	StackTrace() errors.StackTrace
}

type withCause interface {
	Cause() error
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	for _, err := range c.Errors {
		if serverError, found := serverErrorFromCause(err.Err); found {
			log.Warnf("Handler returned error, appErrors.ServerError found in error chain: %d %s",
				serverError.Code, serverError.Error())
			log.Infof("Error: %+v", err.Err)
			//if stackErr, ok := err.Err.(stackError); ok {
			//	log.Infof("Error stacktrace: %+v", err.Err)
			//}
			c.JSON(serverError.Code, gin.H{"error": serverError.Message})
			return
		}
	}

	if len(c.Errors) == 1 {
		log.WithError(c.Errors.Last().Err).Errorf(
			"Handler returned error, appErrors.ServerError not found in error chain, responding with 500")
		if stackErr, ok := c.Errors.Last().Err.(stackError); ok {
			log.Infof("Error stacktrace: %+v", stackErr.StackTrace())
		}
		c.JSON(http.StatusInternalServerError, c.Errors.Last().JSON())
		return
	}
	log.WithField("errors", c.Errors.JSON()).Errorln(
		"Handler returned multiple errors, appErrors.ServerError not found in error chain, responding with 500")
	c.JSON(http.StatusInternalServerError, c.Errors.JSON())
}

func serverErrorFromCause(err error) (serverError *appErrors.ServerError, found bool) {
	for err != nil {
		if e, ok := err.(*appErrors.ServerError); ok {
			return e, true
		}
		err = errors.Unwrap(err)
	}
	return nil, false
}
