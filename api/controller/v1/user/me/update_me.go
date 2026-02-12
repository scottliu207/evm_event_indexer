package me

import (
	"net/http"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	UpdateMeReq struct {
		Password string `json:"password" binding:"required,min=8"`
	}
)

// UpdateMe updates current authenticated user's account and/or password.
func UpdateMe(c *gin.Context) {
	res := new(GetMeRes)
	c.Set(middleware.CtxResponse, res)

	userID, err := getMyID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req UpdateMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if req.Password == "" {
		c.Error(errors.ErrApiInvalidParam.New("account or password is required"))
		return
	}

	user, err := service.UpdateUser(c.Request.Context(), userID, req.Password)
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
