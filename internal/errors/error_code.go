package errors

import (
	"net/http"
)

type Err struct {
	HTTPCode  int
	ErrorCode int
	Message   string
}

// Error implements error.
func (e Err) Error() string {
	return e.Message
}

// Code returns business error code.
func (e Err) Code() int {
	return e.ErrorCode
}

// Define error codes here
var (
	// general error
	API_INVALID_PARAM = Err{HTTPCode: http.StatusBadRequest, ErrorCode: 1000, Message: "invalid api parameter"}
	API_TIMEOUT       = Err{HTTPCode: http.StatusRequestTimeout, ErrorCode: 1001, Message: "api timeout"}

	// account/authorization error
	ACCOUNT_EXISTS    = Err{HTTPCode: http.StatusConflict, ErrorCode: 2000, Message: "account already exists"}
	ACCOUNT_NOT_FOUND = Err{HTTPCode: http.StatusNotFound, ErrorCode: 2001, Message: "account not found"}

	// server error
	INTERNAL_SERVER_ERROR = Err{HTTPCode: http.StatusInternalServerError, ErrorCode: 3000, Message: "something went wrong"}
)
