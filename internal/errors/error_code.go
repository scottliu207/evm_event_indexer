package errors

import (
	"errors"
	"net/http"
	"strings"
)

type Err struct {
	HTTPCode  int
	ErrorCode int
	Message   string
	stack     error
}

// Error implements error.
func (e Err) Error() string {
	return e.Message
}

// Code returns business error code.
func (e Err) Code() int {
	return e.ErrorCode
}

func (e Err) New(AdditionTxt ...string) error {
	msg := make([]string, 0, len(AdditionTxt)+1)
	msg = append(msg, e.Message)
	msg = append(msg, AdditionTxt...)
	return &Err{HTTPCode: e.HTTPCode, ErrorCode: e.ErrorCode, Message: strings.Join(msg, ", ")}
}

func (e Err) Wrap(err error, AdditionTxt ...string) error {

	msg := make([]string, 0, len(AdditionTxt)+1)
	msg = append(msg, e.Message)
	msg = append(msg, AdditionTxt...)
	return &Err{
		HTTPCode:  e.HTTPCode,
		ErrorCode: e.ErrorCode,
		Message:   strings.Join(msg, ": "),
		stack:     err,
	}
}

func (e Err) Unwrap() error {
	return e.stack
}

// Chain returns a string of the error chain with stacked errors concatenated by ", "
func Chain(err error) string {
	parts := []string{}
	for err != nil {
		parts = append(parts, err.Error())
		err = errors.Unwrap(err)
	}

	return strings.Join(parts, ", ")
}

// Define error codes here
var (
	// general error
	ErrApiInvalidParam = Err{HTTPCode: http.StatusBadRequest, ErrorCode: 1000, Message: "invalid api parameter"}
	ErrApiTimeout      = Err{HTTPCode: http.StatusRequestTimeout, ErrorCode: 1001, Message: "api timeout"}

	// account/authorization error
	ErrAccountAlreadyExists = Err{HTTPCode: http.StatusConflict, ErrorCode: 2000, Message: "account already exists"}
	ErrInvalidCredentials   = Err{HTTPCode: http.StatusUnauthorized, ErrorCode: 2001, Message: "invalid credentials"}

	// server error
	ErrInternalServerError = Err{HTTPCode: http.StatusInternalServerError, ErrorCode: 3000, Message: "something went wrong"}
)
