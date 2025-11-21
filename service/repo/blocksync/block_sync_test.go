package blocksync_test

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/db"
	"evm_event_indexer/service/model"

	"evm_event_indexer/service/repo/blocksync"

	"evm_event_indexer/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func Test_TxUpsertBlockSync(t *testing.T) {
	mysql := db.GetMysql(db.EVENT_DB)

	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	err := utils.NewTx(mysql).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return blocksync.TxUpsertBlockSync(ctx, tx, &model.BlockSync{
			Address:                addr,
			LastSyncNumber:         10,
			LastSyncTimestamp:      time.Now(),
			LastFinalizedNumber:    1,
			LastFinalizedTimestamp: time.Now(),
			UpdatedAt:              time.Now(),
		})
	})
	assert.NoError(t, err)

	res, err := blocksync.GetBlockSync(ctx, db.GetMysql(db.EVENT_DB), addr)
	assert.NoError(t, err)
	assert.Equal(t, addr, res.Address)
	assert.Equal(t, uint64(10), res.LastSyncNumber)
	assert.Equal(t, uint64(1), res.LastFinalizedNumber)
}
