package middleware

import (
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"
	"evm_event_indexer/utils/hashing"

	"github.com/gin-gonic/gin"
)

// CSRFProtection validates CSRF token for requests.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {

		rtCookie, err := c.Cookie(CookieNameRefreshToken)
		if err != nil || rtCookie == "" {
			c.Error(errors.ErrInvalidCredentials.New("refresh token is empty"))
			c.Abort()
			return
		}

		// verify csrf token
		sessionData, err := service.VerifyCSRFToken(c.Request.Context(), c.GetHeader(HeaderNameCSRFToken))
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		if sessionData.HashedRT != hashing.Sha256([]byte(rtCookie)) {
			c.Error(errors.ErrInvalidCredentials.New("invalid refresh token"))
			c.Abort()
			return
		}

		c.Next()
	}
}
