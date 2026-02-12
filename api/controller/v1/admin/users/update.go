package users

import (
	"net/http"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	UpdateReq struct {
		UserID   int64           `uri:"user_id" binding:"required,min=1"`
		Password string          `json:"password" binding:"omitempty,min=8"`
		Status   enum.UserStatus `json:"status" binding:"omitempty,oneof=1 2"`
	}
)

func Update(c *gin.Context) {
	res := new(GetRes)
	c.Set(middleware.CtxResponse, res)

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	user, err := service.UpdateUserByAdmin(c.Request.Context(), req.UserID, req.Password, req.Status)
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

	c.Status(http.StatusOK)
}
