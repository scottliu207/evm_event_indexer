package middleware

import (
	"evm_event_indexer/api/protocol"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, protocol.Response{
		Code:    404,
		Message: "API Not Found",
		Result: struct {
		}{},
	})
}
