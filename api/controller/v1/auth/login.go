package auth

import (
	"log/slog"
	"net/http"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"
	"evm_event_indexer/service/repo/session"

	"github.com/gin-gonic/gin"
)

type (
	LoginReq struct {
		Account  string `json:"account" binding:"required,max=50"`
		Password string `json:"password" binding:"required"`
	}

	LoginRes struct {
		AccessToken string `json:"access_token"`
		CSRFToken   string `json:"csrf_token"`
	}
)

// Login logs in the user and returns the access token in body and refresh token in cookie
func Login(c *gin.Context) {
	res := new(LoginRes)
	c.Set(middleware.CtxResponse, res)

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	user, err := service.VerifyUserPassword(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	// delete old refresh token if exists
	rtCookie, _ := c.Cookie(middleware.CookieNameRefreshToken)
	if rtCookie != "" {
		if err := service.DeleteUserRT(c.Request.Context(), rtCookie); err != nil {
			slog.Error("failed to delete refresh token", slog.Any("error", err), slog.Any("userID", user.ID))
		}
	}

	at, err := service.CreateUserAT(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(errors.ErrInternalServerError.Wrap(err, "failed to create access token"))
		return
	}

	rt, err := service.CreateUserRT(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(errors.ErrInternalServerError.Wrap(err, "failed to create refresh token"))
		return
	}

	*res = LoginRes{
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
	// cross site
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(middleware.CookieNameRefreshToken, rt, int(config.Get().Session.RTExpiration.Seconds()), "/", "", true, true)
	c.SetCookie(middleware.CookieNameCSRFToken, csrfToken, int(config.Get().Session.RTExpiration.Seconds()), "/", "", true, false)
}
