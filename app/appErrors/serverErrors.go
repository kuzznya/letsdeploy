package appErrors

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type ServerError struct {
	cause   error
	Code    int
	Message string
}

func (s *ServerError) Error() string {
	return s.Message
}

func (s *ServerError) Cause() error {
	return s.cause
}

// Format formats ServerError the same way as github.com/pkg/errors does
func (s *ServerError) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') && s.Cause() != nil {
			fmt.Fprintf(state, "%+v\n", s.Cause())
			io.WriteString(state, s.Error())
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(state, s.Error())
	}
}

func Wrap(err error, code int, message string) error {
	return errors.WithStack(&ServerError{cause: err, Code: code, Message: message})
}

func BadRequest(message string) error {
	return errors.WithStack(&ServerError{Code: http.StatusBadRequest, Message: message})
}

func BadRequestWrap(err error, message string) error {
	return Wrap(err, http.StatusBadRequest, message)
}

func Unauthorized(message string) error {
	return errors.WithStack(&ServerError{Code: http.StatusUnauthorized, Message: message})
}

func UnauthorizedWrap(err error, message string) error {
	return Wrap(err, http.StatusUnauthorized, message)
}

func Forbidden(message string) error {
	return errors.WithStack(&ServerError{Code: http.StatusForbidden, Message: message})
}

func ForbiddenWrap(err error, message string) error {
	return Wrap(err, http.StatusForbidden, message)
}

func NotFound(message string) error {
	return errors.WithStack(&ServerError{Code: http.StatusNotFound, Message: message})
}

func NotFoundWrap(err error, message string) error {
	return Wrap(err, http.StatusNotFound, message)
}

func InternalServerError(message string) error {
	return errors.WithStack(&ServerError{Code: http.StatusInternalServerError, Message: message})
}

func InternalServerErrorWrap(err error, message string) error {
	return Wrap(err, http.StatusInternalServerError, message)
}
