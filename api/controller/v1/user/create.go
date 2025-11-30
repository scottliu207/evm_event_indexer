package user

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"evm_event_indexer/api/protocol"
	internalCnf "evm_event_indexer/internal/config"
	"evm_event_indexer/internal/enum"
	internalStorage "evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	userRepo "evm_event_indexer/service/repo/user"
	"evm_event_indexer/utils"
	"evm_event_indexer/utils/hashing"

	"github.com/gin-gonic/gin"
)

type (
	CreateUserReq struct {
		Account  string `json:"Account" binding:"required,min=3,max=20"`
		Password string `json:"Password" binding:"required,min=8"`
	}

	CreateUserRes struct {
		ID        int64           `json:"ID"`
		Account   string          `json:"Account"`
		Role      enum.UserRole   `json:"Role"`
		Status    enum.UserStatus `json:"Status"`
		CreatedAt time.Time       `json:"CreatedAt"`
	}
)

// Create handles creating a new user with Argon2 password hashing.
// It only supports creating users with default role (1) and enabled status (1).
func Create(c *gin.Context) {
	res := &protocol.Response{}

	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	cnf := internalCnf.Get()

	argonOpt := &hashing.Argon2Opt{
		Time:    cnf.Argon2.Time,
		Memory:  cnf.Argon2.Memory,
		Threads: cnf.Argon2.Threads,
		KeyLen:  cnf.Argon2.KeyLen,
	}

	hasher := hashing.NewArgon2(argonOpt)
	hashB64, saltB64, err := hasher.Hashing(req.Password)
	if err != nil {
		c.Error(err)
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

	dbm := internalStorage.GetMysql(cnf.MySQL.EventDBM.DBName)
	txObj := utils.NewTx(dbm)

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
		c.Error(err)
		return
	}

	res.Result = &CreateUserRes{
		ID:        newUser.ID,
		Account:   newUser.Account,
		Role:      newUser.Role,
		Status:    newUser.Status,
		CreatedAt: newUser.CreatedAt,
	}
	c.Status(http.StatusCreated)
}
