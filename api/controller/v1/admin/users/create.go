package users

import (
	"net/http"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	CreateReq struct {
		Account  string `json:"account" binding:"required,min=3,max=20"`
		Password string `json:"password" binding:"required,min=8"`
	}
)

func Create(c *gin.Context) {
	res := new(GetRes)
	c.Set(middleware.CtxResponse, res)

	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	exists, err := service.GetUserByAccount(c.Request.Context(), req.Account)
	if err != nil {
		c.Error(err)
		return
	}
	if exists != nil {
		c.Error(errors.ErrAccountAlreadyExists.New())
		return
	}

	user, err := service.CreateUserByAdmin(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	*res = GetRes{
		ID:        user.ID,
		Account:   user.Account,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.Status(http.StatusCreated)
}
