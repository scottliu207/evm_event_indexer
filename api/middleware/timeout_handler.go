package middleware

import (
	"evm_event_indexer/api/protocol"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"net/http"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

// api timeout handler
func TimeoutHandler() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(config.Get().API.Timeout),
		timeout.WithResponse(timeoutRes),
	)
}

func timeoutRes(c *gin.Context) {
	c.JSON(http.StatusRequestTimeout, protocol.Response{
		Code:    errors.ErrApiTimeout.ErrorCode,
		Message: errors.ErrApiTimeout.Message,
	})
}
