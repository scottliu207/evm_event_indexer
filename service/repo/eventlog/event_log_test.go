package eventlog_test

import (
	"context"
	"database/sql"
	"os"

	internalCnf "evm_event_indexer/internal/config"
	internalStorage "evm_event_indexer/internal/storage"
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
	internalCnf.LoadConfig("../../../config/config.yaml")
	internalStorage.InitDB()

	os.Exit(m.Run())
}

func Test_LogRepo(t *testing.T) {
	internalCnf.LoadConfig("../../../config/config.yaml")
	cnf := internalCnf.Get()
	db := internalStorage.GetMysql(cnf.MySQL.EventDBS.DBName)
	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	err := utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
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
	total, err := eventlog.GetTotal(ctx, param)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	logs, err := eventlog.GetLogs(ctx, param)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Equal(t, addr, logs[0].Address)

}
