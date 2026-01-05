package auth

import (
	"net/http"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"
	"evm_event_indexer/service/repo/session"

	"github.com/gin-gonic/gin"
)

type (
	RotateTokenReq struct {
	}

	RotateTokenRes struct {
		AccessToken string `json:"access_token"`
		CSRFToken   string `json:"csrf_token"`
	}
)

// RotateToken rotates the user access token and refresh token
func RotateToken(c *gin.Context) {
	res := new(RotateTokenRes)
	c.Set(middleware.CtxResponse, res)

	// get refresh token from cookie
	rtCookie, err := c.Cookie("refresh_token")
	if err != nil && err != http.ErrNoCookie {
		c.Error(errors.ErrInvalidCredentials.New("refresh token is not found"))
		return
	}

	if rtCookie == "" {
		c.Error(errors.ErrInvalidCredentials.New("refresh token is not found"))
		return
	}

	userID, err := service.GetUserIDByRT(c.Request.Context(), rtCookie)
	if err != nil {
		c.Error(err)
		return
	}

	if err := service.DeleteUserRT(c.Request.Context(), rtCookie); err != nil {
		c.Error(err)
		return
	}

	at, err := service.CreateUserAT(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	rt, err := service.CreateUserRT(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	*res = RotateTokenRes{
		AccessToken: at,
	}

	csrfToken, err := session.NewCSRFToken(rt)
	if err != nil {
		c.Error(err)
		return
	}
	res.CSRFToken = csrfToken

	// set refresh token & csrf token cookie
	c.Status(http.StatusOK)
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("refresh_token", rt, int(config.Get().Session.RTExpiration.Seconds()), "/", "", true, true)
	c.SetCookie(middleware.CSRFCookieName, csrfToken, int(config.Get().Session.RTExpiration.Seconds()), "/", "", true, false)
}
