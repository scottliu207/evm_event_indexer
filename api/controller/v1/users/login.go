package users

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
		Account  string `json:"account" binding:"required,max=20"`
		Password string `json:"password" binding:"required"`
	}

	LoginRes struct {
		AccessToken string `json:"access_token"`
	}
)

// User login. Returns access token in body and refresh token in cookie.
func Login(c *gin.Context) {
	res := new(LoginRes)
	c.Set(middleware.CtxResponse, res)

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	cnf := config.Get()

	user, err := service.VerifyUserPassword(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		c.Error(errors.ErrInternalServerError.Wrap(err, "failed to verify user password"))
		return
	}

	if user == nil {
		c.Error(errors.ErrInvalidCredentials.New())
		return
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

	// set refresh token cookie
	c.Status(http.StatusOK)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", rt, int(cnf.Session.RTExpiration.Seconds()), "/", "", true, true)
}
