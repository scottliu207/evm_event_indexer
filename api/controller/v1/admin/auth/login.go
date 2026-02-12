package auth

import (
	"net/http"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	LoginReq struct {
		Account  string `json:"account" binding:"required,max=50"`
		Password string `json:"password" binding:"required"`
	}

	LoginRes struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
		CSRFToken   string `json:"csrf_token"`
	}
)

func Login(c *gin.Context) {
	res := new(LoginRes)
	c.Set(middleware.CtxResponse, res)

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	admin, err := service.VerifyAdminPassword(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	s, err := service.CreateAdminSession(c.Request.Context(), admin.ID)
	if err != nil {
		c.Error(errors.ErrInternalServerError.Wrap(err, "failed to create session"))
		return
	}

	*res = LoginRes{
		AccessToken: s.AT,
		ExpiresAt:   s.ATExpiresAt.Unix(),
		CSRFToken:   s.CSRFToken,
	}

	c.Status(http.StatusOK)
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(middleware.CookieNameAdminRefreshToken, s.RT, int(config.Get().Session.SessionExpiration.Seconds()), "/", "", true, true)
}
