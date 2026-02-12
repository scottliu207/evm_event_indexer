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
	RotateTokenReq struct {
	}

	RotateTokenRes struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"` // timestamp of access token expiration
		CSRFToken   string `json:"csrf_token"`
	}
)

// RotateToken rotates the user access token and refresh token
//
//	@Summary		Refresh access token
//	@Description	Rotate access and refresh tokens using the refresh_token cookie. Returns new tokens and sets a new refresh_token cookie.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	protocol.Response{result=RotateTokenRes}
//	@Failure		401	{object}	protocol.Response
//	@Router			/v1/auth/refresh [post]
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

	// get user id by refresh token
	userID, err := service.GetUserIDByRT(c.Request.Context(), rtCookie)
	if err != nil {
		c.Error(err)
		return
	}

	// revoke old session and create new session
	session, err := service.CreateSession(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	*res = RotateTokenRes{
		AccessToken: session.AT,
		ExpiresAt:   session.ATExpiresAt.Unix(),
		CSRFToken:   session.CSRFToken,
	}

	c.Status(http.StatusOK)
	c.SetSameSite(http.SameSiteNoneMode)
	// set new refresh token
	c.SetCookie(middleware.CookieNameRefreshToken, session.RT, int(config.Get().Session.SessionExpiration.Seconds()), "/", "", true, true)
}
