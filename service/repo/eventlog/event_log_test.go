package eventlog_test

import (
	"context"
	"database/sql"

	internalCnf "evm_event_indexer/internal/config"
	internalStorage "evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func Test_TxInsertLog(t *testing.T) {
	cnf := internalCnf.Get()
	db := internalStorage.GetMysql(cnf.EventDB)
	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	err := utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return eventlog.TxUpsertLog(ctx, tx, &model.Log{
			Address:     addr,
			BlockHash:   common.Hash{}.String(),
			BlockNumber: 0,
			Topics: &model.Topics{
				"0x123",
				"0x456",
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

	log, err := eventlog.GetEventLog(ctx, db, 35)
	spew.Dump(log)
	assert.NoError(t, err)
	assert.Equal(t, addr, log.Address)
}
