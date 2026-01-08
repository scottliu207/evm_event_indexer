package middleware

import (
	"crypto/hmac"
	"strings"

	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service/repo/session"

	"github.com/gin-gonic/gin"
)

// CSRFProtection validates CSRF token for requests.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {

		// get rt from cookie, if not found, skip handling csrf
		// csrf token protection is only needed when refresh token has been sent along with the request
		rt, err := c.Cookie("refresh_token")
		if err != nil || rt == "" {
			c.Next()
			return
		}

		csrfCookie, _ := c.Cookie(CookieNameCSRFToken)
		csrfHeader := strings.TrimSpace(c.GetHeader("X-CSRF-Token"))
		if csrfCookie == "" || csrfHeader == "" {
			c.Error(errors.ErrCSRFTokenInvalid.New("csrf token is required"))
			c.Abort()
			return
		}

		if !hmac.Equal([]byte(csrfCookie), []byte(csrfHeader)) {
			c.Error(errors.ErrCSRFTokenInvalid.New("csrf token mismatch"))
			c.Abort()
			return
		}

		expected, err := session.NewCSRFToken(rt)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if !hmac.Equal([]byte(expected), []byte(csrfHeader)) {
			c.Error(errors.ErrCSRFTokenInvalid.New("csrf token invalid"))
			c.Abort()
			return
		}

		c.Next()
	}
}
