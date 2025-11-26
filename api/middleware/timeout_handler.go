package middleware

import (
	"evm_event_indexer/api/protocol"
	internalCnf "evm_event_indexer/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/heshuosg/system/cons/module/er.git"
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
				er.APITimeout100002.HttpCode,
				&protocol.Response{
					Code:    er.APITimeout100002.Code,
					Message: er.APITimeout100002.Message,
				})
		case <-finish:
		}
	}
}
