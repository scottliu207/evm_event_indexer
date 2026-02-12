package auth

import (
	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	LogoutReq struct {
	}

	LogoutRes struct {
	}
)

// Logout logs out the user by deleting the refresh token
func Logout(c *gin.Context) {
	res := new(LogoutRes)
	c.Set(middleware.CtxResponse, res)

	var req LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	userID := c.GetInt64(middleware.CtxUserID)
	if userID == 0 {
		c.Error(errors.ErrInvalidCredentials.New("user id not found"))
		return
	}

	// revoke user session
	if err := service.RevokeUserSession(c.Request.Context(), userID); err != nil {
		c.Error(err)
		return
	}

	// delete refresh token cookie
	c.SetCookie(middleware.CookieNameRefreshToken, "", -1, "/", "", false, true)
}
