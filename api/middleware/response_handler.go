package middleware

import (
	"encoding/json"
	"errors"
	"io"

	"evm_event_indexer/api/protocol"
	internalErr "evm_event_indexer/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const CTX_RESPONSE = "response"

// ResponseHandler automatically handles API responses and errors.
func ResponseHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// If response already written, skip further processing
		if c.Writer.Written() && !c.IsAborted() {
			return
		}

		res := &protocol.Response{
			Result: new(struct{}),
		}

		// Check for errors after processing, only handle the first error
		last := c.Errors.Last()

		// no error occurred, return success
		if last == nil {
			res.Code = 0
			res.Message = "success"

			// return empty object when result is nil to maintain consistent JSON structure
			result, ok := c.Get(CTX_RESPONSE)
			if ok {
				res.Result = result
			}

			c.JSON(c.Writer.Status(), res)
			return
		}

		var err *internalErr.Err
		switch {
		case errors.As(last.Err, &err):
			res.Code = err.ErrorCode
			res.Message = err.Message
			c.JSON(err.HTTPCode, res)
		case last.Type == gin.ErrorTypeBind || isBindErr(last.Err):
			res.Code = internalErr.API_INVALID_PARAM.ErrorCode
			res.Message = internalErr.API_INVALID_PARAM.Message
			c.JSON(internalErr.API_INVALID_PARAM.HTTPCode, res)
		default:
			res.Code = internalErr.INTERNAL_SERVER_ERROR.ErrorCode
			res.Message = internalErr.INTERNAL_SERVER_ERROR.Message
			c.JSON(internalErr.INTERNAL_SERVER_ERROR.HTTPCode, res)
		}
	}
}

func isBindErr(err error) bool {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return true
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return true
	}
	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		return true
	}
	return errors.Is(err, io.EOF)
}
