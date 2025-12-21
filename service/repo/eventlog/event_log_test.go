package eventlog_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/internal/testutil"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func TestMain(m *testing.M) {
	testutil.SetupTestConfig()
	dbManager := storage.Forge()
	if err := dbManager.Init(); err != nil {
		panic(fmt.Sprintf("failed to init database: %s\n", err))
	}

	code := m.Run()
	dbManager.Shutdown()
	os.Exit(code)
}

func Test_LogRepo(t *testing.T) {
	db, err := storage.GetMySQL(config.EventDBM)
	if err != nil {
		t.Fatalf("failed to get mysql: %s\n", err)
	}

	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	err = utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return eventlog.TxUpsertLog(ctx, tx, &model.Log{
			Address:     addr,
			BlockHash:   common.Hash{}.String(),
			BlockNumber: 0,
			Topics: &model.Topics{
				common.HexToHash("0x123"),
				common.HexToHash("0x456"),
			},
			TxIndex:        1,
			LogIndex:       2,
			TxHash:         common.Hash{}.String(),
			Data:           []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			BlockTimestamp: time.Now(),
			CreatedAt:      time.Now(),
		})
	})
	assert.NoError(t, err)

	param := &eventlog.GetLogParam{
		Address:    addr,
		StartTime:  time.Now().Add(-time.Hour),
		EndTime:    time.Now(),
		Pagination: &model.Pagination{Page: 1, Size: 10},
	}
	total, err := eventlog.GetTotal(ctx, db, param)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	logs, err := eventlog.GetLogs(ctx, db, param)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Equal(t, addr, logs[0].Address)
}
