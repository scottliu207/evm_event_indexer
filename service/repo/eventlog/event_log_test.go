package eventlog_test

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/db"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func Test_TxInsertLog(t *testing.T) {
	mysql := db.GetMysql(db.EVENT_DB)
	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	err := utils.NewTx(mysql).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return eventlog.TxUpsertLog(ctx, tx, &model.Log{
			Address:     addr,
			BlockHash:   common.Hash{}.String(),
			BlockNumber: 0,
			Topics: &model.Topics{
				"0x123",
				"0x456",
			},
			TxIndex:        0,
			LogIndex:       0,
			TxHash:         common.Hash{}.String(),
			Data:           []byte{},
			BlockTimestamp: time.Now(),
		})
	})
	assert.NoError(t, err)
}
