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
		UserID   int64           `uri:"user_id" json:"-" binding:"required,min=1"`
		Password string          `json:"password" binding:"omitempty,min=8"`
		Status   enum.UserStatus `json:"status" binding:"omitempty,oneof=1 2"`
	}
)

// Update updates a user's password and/or status.
//
//	@Summary		Update user
//	@Description	Update a user's password and/or status by ID (admin only).
//	@Tags			Admin Users
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		int			true	"User ID"
//	@Param			request	body		UpdateReq	true	"Fields to update"
//	@Success		200		{object}	protocol.Response{result=GetRes}
//	@Failure		400		{object}	protocol.Response
//	@Failure		401		{object}	protocol.Response
//	@Failure		404		{object}	protocol.Response
//	@Security		AdminBearerAuth
//	@Router			/v1/admin/users/{user_id} [put]
func Update(c *gin.Context) {
	res := new(GetRes)
	c.Set(middleware.CtxResponse, res)

	var req UpdateReq
	if err := c.ShouldBindUri(&req); err != nil {
		c.Error(err)
		return
	}

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
