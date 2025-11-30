package middleware

import (
	"evm_event_indexer/api/protocol"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFoundHandler handles 404 Not Found requests and returns a standardized response.
func NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, protocol.Response{
		Code:    http.StatusNotFound,
		Message: "api not found",
		Result:  struct{}{},
	})
}
