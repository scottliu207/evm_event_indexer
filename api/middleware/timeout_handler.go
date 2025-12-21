package middleware

import (
	"evm_event_indexer/api/protocol"
	internalCnf "evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"time"

	"github.com/gin-gonic/gin"
)

// api timeout handler
func TimeoutHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		cnf := internalCnf.Get()
		finish := make(chan struct{})
		panicChan := make(chan any, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			finish <- struct{}{}
		}()

		select {
		case p := <-panicChan:
			panic(p)
		case <-time.After(cnf.API.Timeout):
			c.AbortWithStatusJSON(
				errors.ErrApiTimeout.HTTPCode,
				&protocol.Response{
					Code:    errors.ErrApiTimeout.ErrorCode,
					Message: errors.ErrApiTimeout.Message,
				})
		case <-finish:
		}
	}
}
