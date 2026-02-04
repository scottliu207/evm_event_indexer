package auth_test

import (
	"context"
	"database/sql"
	"evm_event_indexer/api"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/internal/testutil"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/user"
	"evm_event_indexer/utils"
	"evm_event_indexer/utils/hashing"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	ctx    = context.Background()
)

var (
	testAccount  string
	testPassword = "password123"
)

func TestMain(m *testing.M) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()

	dbManager := storage.Forge()
	if err := dbManager.Init(); err != nil {
		panic(fmt.Sprintf("failed to init database: %s\n", err))
	}

	// Setup router (same middleware chain as production server)
	server := api.NewServer()
	router = server.Handler.(*gin.Engine)

	// Create test user
	// account_db.user.account is VARCHAR(20); keep it short to avoid truncation.
	testAccount = fmt.Sprintf("tlu_%d", time.Now().UnixNano()%1_000_000_000)
	if err := createTestUser(); err != nil {
		panic(fmt.Sprintf("failed to create test user: %s\n", err))
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestUser()
	dbManager.Shutdown()
	os.Exit(code)
}

func createTestUser() error {
	db, err := storage.GetMySQL(config.AccountDBM)
	if err != nil {
		return err
	}

	opt := &hashing.Argon2Opt{
		Time:    config.Get().Argon2.Time,
		Memory:  config.Get().Argon2.Memory,
		Threads: config.Get().Argon2.Threads,
		KeyLen:  config.Get().Argon2.KeyLen,
	}

	hasher := hashing.NewArgon2(opt)
	pwdB64, saltB64, err := hasher.Hashing(testPassword)
	if err != nil {
		return err
	}

	return utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := user.TxInsertUser(ctx, tx, &model.User{
			Account:  testAccount,
			Status:   enum.UserStatusEnabled,
			Password: pwdB64,
			AuthMeta: &model.AuthMeta{
				Salt:    saltB64,
				Memory:  opt.Memory,
				Time:    opt.Time,
				Threads: opt.Threads,
				KeyLen:  opt.KeyLen,
			},
			CreatedAt: time.Now(),
		})
		return err
	})
}

func cleanupTestUser() {
	db, err := storage.GetMySQL(config.AccountDBM)
	if err != nil {
		return
	}

	users, err := user.GetUsers(ctx, db, &user.GetUserFilter{
		Accounts: []string{testAccount},
	})
	if err != nil || len(users) == 0 {
		return
	}

	_ = utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return user.TxDeleteUser(ctx, tx, users[0].ID)
	})
}
