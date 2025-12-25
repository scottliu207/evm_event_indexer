package auth

import (
	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"
	"net/http"

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

	// get refresh token from cookie
	rtCookie, err := c.Cookie("refresh_token")
	if err != nil && err != http.ErrNoCookie {
		c.Error(errors.ErrInvalidCredentials.New("refresh token is not found"))
		return
	}

	// if refresh token is not found, return directly
	if rtCookie == "" {
		return
	}

	// delete refresh token
	if err := service.DeleteUserRT(c.Request.Context(), rtCookie); err != nil {
		c.Error(err)
		return
	}

	*res = LogoutRes{}
}
