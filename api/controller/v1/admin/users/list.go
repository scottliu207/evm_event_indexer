package users

import (
	"net/http"
	"time"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/service"
	"evm_event_indexer/service/model"
	userRepo "evm_event_indexer/service/repo/user"

	"github.com/gin-gonic/gin"
)

type (
	ListReq struct {
		Page   uint64 `form:"page" binding:"required,min=1"`
		Size   uint64 `form:"size" binding:"required,min=1,max=100"`
		Status int8   `form:"status" binding:"omitempty"`
	}

	ListRes struct {
		Users []Row `json:"users"`
		Total int64 `json:"total"`
	}

	Row struct {
		ID        int64           `json:"id"`
		Account   string          `json:"account"`
		Status    enum.UserStatus `json:"status"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
	}
)

func List(c *gin.Context) {
	res := &ListRes{
		Users: make([]Row, 0),
	}
	c.Set(middleware.CtxResponse, res)

	var req ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	filter := &userRepo.GetUserFilter{
		Pagination: &model.Pagination{Page: req.Page, Size: req.Size},
	}
	if req.Status != 0 {
		filter.Status = enum.UserStatus(req.Status)
	}

	users, total, err := service.GetUsersByAdmin(c.Request.Context(), filter)
	if err != nil {
		c.Error(err)
		return
	}

	res.Total = total
	res.Users = make([]Row, len(users))
	for i, u := range users {
		res.Users[i] = Row{
			ID:        u.ID,
			Account:   u.Account,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	c.Status(http.StatusOK)
}
