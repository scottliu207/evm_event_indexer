package me

import (
	"net/http"
	"time"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

// UserRes is the response DTO for a single user (excludes password and auth_meta).
type GetMeRes struct {
	ID        int64     `json:"id"`
	Account   string    `json:"account"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetMe returns the current authenticated user profile.
//
//	@Summary		Get current user profile
//	@Description	Returns the profile of the currently authenticated user.
//	@Tags			User
//	@Produce		json
//	@Success		200	{object}	protocol.Response{result=GetMeRes}
//	@Failure		401	{object}	protocol.Response
//	@Security		BearerAuth
//	@Router			/v1/user/me [get]
func GetMe(c *gin.Context) {
	res := new(GetMeRes)
	c.Set(middleware.CtxResponse, res)

	userID, err := getMyID(c)
	if err != nil {
		c.Error(err)
		return
	}

	user, err := service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	*res = GetMeRes{
		ID:        user.ID,
		Account:   user.Account,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.Status(http.StatusOK)
}
