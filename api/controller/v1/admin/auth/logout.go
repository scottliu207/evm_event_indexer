package auth

import (
	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type LogoutRes struct{}

func Logout(c *gin.Context) {
	res := new(LogoutRes)
	c.Set(middleware.CtxResponse, res)

	adminID := c.GetInt64(middleware.CtxAdminID)
	if adminID == 0 {
		c.Error(errors.ErrInvalidCredentials.New("admin id not found"))
		return
	}

	if err := service.RevokeAdminSession(c.Request.Context(), adminID); err != nil {
		c.Error(err)
		return
	}

	c.SetCookie(middleware.CookieNameAdminRefreshToken, "", -1, "/", "", false, true)
}
