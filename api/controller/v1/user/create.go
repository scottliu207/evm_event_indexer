package user

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"evm_event_indexer/api/middleware"
	internalCnf "evm_event_indexer/internal/config"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"

	userRepo "evm_event_indexer/service/repo/user"
	"evm_event_indexer/utils"
	"evm_event_indexer/utils/hashing"

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
		Role      enum.UserRole   `json:"role"`
		Status    enum.UserStatus `json:"status"`
		CreatedAt time.Time       `json:"created_at"`
	}
)

// Create handles creating a new user with Argon2 password hashing.
// It only supports creating users with default role (1) and enabled status (1).
func Create(c *gin.Context) {
	res := new(CreateUserRes)
	c.Set(middleware.CTX_RESPONSE, res)

	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	cnf := internalCnf.Get()

	users, _, err := userRepo.GetUsers(c.Request.Context(), storage.GetMysql(cnf.MySQL.EventDBS.DBName), &userRepo.GetUserFilter{
		Accounts: []string{req.Account},
	})
	if err != nil {
		c.Error(errors.INTERNAL_SERVER_ERROR.Wrap(err, "failed to get user info"))
		return
	}

	if len(users) > 0 {
		c.Error(errors.ACCOUNT_ALREADY_EXISTS.New())
		return
	}

	argonOpt := &hashing.Argon2Opt{
		Time:    cnf.Argon2.Time,
		Memory:  cnf.Argon2.Memory,
		Threads: cnf.Argon2.Threads,
		KeyLen:  cnf.Argon2.KeyLen,
	}

	hasher := hashing.NewArgon2(argonOpt)
	hashB64, saltB64, err := hasher.Hashing(req.Password)
	if err != nil {
		c.Error(errors.INTERNAL_SERVER_ERROR.Wrap(err, "failed to hash password"))
		return
	}

	now := time.Now()
	newUser := &model.User{
		Account:  req.Account,
		Status:   enum.UserStatusEnabled,
		Role:     enum.UserRoleUser,
		Password: hashB64,
		AuthMeta: &model.AuthMeta{
			Salt:    saltB64,
			Memory:  argonOpt.Memory,
			Time:    argonOpt.Time,
			Threads: argonOpt.Threads,
			KeyLen:  argonOpt.KeyLen,
		},
		CreatedAt: now,
	}

	txObj := utils.NewTx(storage.GetMysql(cnf.MySQL.EventDBM.DBName))
	err = txObj.Exec(c.Request.Context(),
		func(ctx context.Context, tx *sql.Tx) error {
			id, err := userRepo.TxInsertUser(ctx, tx, newUser)
			if err != nil {
				return err
			}
			newUser.ID = id
			return nil
		})
	if err != nil {
		c.Error(errors.INTERNAL_SERVER_ERROR.Wrap(err, "failed to create user"))
		return
	}

	*res = CreateUserRes{
		ID:        newUser.ID,
		Account:   newUser.Account,
		Role:      newUser.Role,
		Status:    newUser.Status,
		CreatedAt: newUser.CreatedAt,
	}

	c.Status(http.StatusCreated)
}
