package users

import (
	"net/http"
	"time"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	GetReq struct {
		UserID int64 `uri:"user_id" binding:"required,min=1"`
	}

	GetRes struct {
		ID        int64           `json:"id"`
		Account   string          `json:"account"`
		Status    enum.UserStatus `json:"status"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
	}
)

// Get returns a single user by ID.
//
//	@Summary		Get user
//	@Description	Get a single user by ID (admin only).
//	@Tags			Admin Users
//	@Produce		json
//	@Param			user_id	path		int	true	"User ID"
//	@Success		200		{object}	protocol.Response{result=GetRes}
//	@Failure		401		{object}	protocol.Response
//	@Failure		404		{object}	protocol.Response
//	@Security		AdminBearerAuth
//	@Router			/v1/admin/users/{user_id} [get]
func Get(c *gin.Context) {
	res := new(GetRes)
	c.Set(middleware.CtxResponse, res)

	var req = new(GetReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.Error(err)
		return
	}

	user, err := service.GetUserByIDByAdmin(c.Request.Context(), req.UserID)
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
