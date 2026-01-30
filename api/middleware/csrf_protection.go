package middleware

import (
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

// CSRFProtection validates CSRF token for requests.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {

		// verify csrf token
		if err := service.VerifyCSRFToken(c.Request.Context(), c.GetHeader(HeaderNameCSRFToken)); err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		c.Next()
	}
}
