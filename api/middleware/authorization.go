package middleware

import (
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

// Authorization validates authentication for protected routes
func Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {

		at := c.GetHeader("Authorization")
		if at == "" {
			c.Error(errors.ErrApiInvalidParam.New("access token is required"))
			c.Abort()
			return
		}

		userID, err := service.VerifyUserAT(c.Request.Context(), at)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		c.Set(CtxUserID, userID)
		c.Next()
	}
}
