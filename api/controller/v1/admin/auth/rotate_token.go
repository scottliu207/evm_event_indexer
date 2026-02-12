package auth

import (
	"net/http"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type RotateTokenRes struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
	CSRFToken   string `json:"csrf_token"`
}

func RotateToken(c *gin.Context) {
	res := new(RotateTokenRes)
	c.Set(middleware.CtxResponse, res)

	rtCookie, err := c.Cookie(middleware.CookieNameAdminRefreshToken)
	if err != nil && err != http.ErrNoCookie {
		c.Error(errors.ErrInvalidCredentials.New("refresh token is not found"))
		return
	}
	if rtCookie == "" {
		c.Error(errors.ErrInvalidCredentials.New("refresh token is not found"))
		return
	}

	adminID, err := service.GetAdminIDByRT(c.Request.Context(), rtCookie)
	if err != nil {
		c.Error(err)
		return
	}

	s, err := service.CreateAdminSession(c.Request.Context(), adminID)
	if err != nil {
		c.Error(err)
		return
	}

	*res = RotateTokenRes{
		AccessToken: s.AT,
		ExpiresAt:   s.ATExpiresAt.Unix(),
		CSRFToken:   s.CSRFToken,
	}

	c.Status(http.StatusOK)
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(middleware.CookieNameAdminRefreshToken, s.RT, int(config.Get().Session.SessionExpiration.Seconds()), "/", "", true, true)
}
