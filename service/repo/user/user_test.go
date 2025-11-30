package user_test

import (
	"context"
	"database/sql"
	"os"

	internalCnf "evm_event_indexer/internal/config"
	"evm_event_indexer/internal/enum"
	internalStorage "evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/user"
	"evm_event_indexer/utils"
	"evm_event_indexer/utils/hashing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func TestMain(m *testing.M) {
	internalCnf.LoadConfig("../../../config/config.yaml")
	internalStorage.InitDB()

	os.Exit(m.Run())
}

func Test_User(t *testing.T) {

	const testUser = "test_user"

	opt := &hashing.Argon2Opt{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 2,
		KeyLen:  32,
	}

	a := hashing.NewArgon2(opt)
	pwdB64, saltB64, err := a.Hashing("password123")
	assert.NoError(t, err)

	cnf := internalCnf.Get()
	db := internalStorage.GetMysql(cnf.MySQL.EventDBS.DBName)
	txObj := utils.NewTx(db)
	err = txObj.Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			_, err := user.TxInsertUser(ctx, tx, &model.User{
				Account:  testUser,
				Status:   enum.UserStatusEnabled,
				Role:     enum.UserRoleUser,
				Password: pwdB64,
				AuthMeta: &model.AuthMeta{
					Salt:    saltB64,
					Memory:  uint32(opt.Memory),
					Time:    uint32(opt.Time),
					Threads: uint8(opt.Threads),
					KeyLen:  uint32(opt.KeyLen),
				},
				CreatedAt: time.Now(),
			})
			return err
		})
	assert.NoError(t, err)

	users, _, err := user.GetUsers(ctx, db, &user.GetUserFilter{
		Accounts: []string{testUser},
		Status:   enum.UserStatusEnabled,
		Role:     enum.UserRoleUser,
	})
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, testUser, users[0].Account)
	assert.Equal(t, enum.UserStatusEnabled, users[0].Status)
	assert.Equal(t, enum.UserRoleUser, users[0].Role)
	t.Logf("user retrieved: %v", users[0])

	err = txObj.Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			return user.TxDeleteUser(ctx, tx, users[0].ID)
		})
	assert.NoError(t, err)
}
