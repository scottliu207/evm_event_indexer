package users

import (
	"net/http"
	"time"

	"evm_event_indexer/api/middleware"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/service"

	"github.com/gin-gonic/gin"
)

type (
	CreateUserReq struct {
		Account  string `json:"account" binding:"required,min=3,max=20"`
		Password string `json:"password" binding:"required,min=8"`
	}

	CreateUserRes struct {
		ID        int64           `json:"id"`
		Account   string          `json:"account"`
		Status    enum.UserStatus `json:"status"`
		CreatedAt time.Time       `json:"created_at"`
	}
)

// Create creates a new user with Argon2 password hashing
// currently not used
func Create(c *gin.Context) {
	res := new(CreateUserRes)
	c.Set(middleware.CtxResponse, res)

	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	user, err := service.GetUserByAccount(c.Request.Context(), req.Account)
	if err != nil {
		c.Error(err)
		return
	}

	if user != nil {
		c.Error(errors.ErrAccountAlreadyExists.New())
		return
	}

	nUser, err := service.InsertUser(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	*res = CreateUserRes{
		ID:        nUser.ID,
		Account:   nUser.Account,
		Status:    nUser.Status,
		CreatedAt: nUser.CreatedAt,
	}

	c.Status(http.StatusCreated)
}
