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
		ExpiresAt   int64  `json:"expires_at"` // timestamp of access token expiration
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

	// revoke old session and create new session
	session, err := service.CreateSession(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(errors.ErrInternalServerError.Wrap(err, "failed to create session"))
		return
	}

	*res = LoginRes{
		AccessToken: session.AT,
		ExpiresAt:   session.ATExpiresAt.Unix(),
		CSRFToken:   session.CSRFToken,
	}

	c.Status(http.StatusOK)
	// cross site
	c.SetSameSite(http.SameSiteNoneMode)
	// set refresh token
	c.SetCookie(middleware.CookieNameRefreshToken, session.RT, int(config.Get().Session.SessionExpiration.Seconds()), "/", "", true, true)
}
