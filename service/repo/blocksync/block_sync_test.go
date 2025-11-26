package blocksync_test

import (
	"context"
	"database/sql"
	internalCnf "evm_event_indexer/internal/config"
	internalStorage "evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"

	"evm_event_indexer/service/repo/blocksync"

	"evm_event_indexer/utils"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func Test_TxUpsertBlock(t *testing.T) {
	cnf := internalCnf.Get()
	db := internalStorage.GetMysql(cnf.EventDB)

	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	err := utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
			Address:        addr,
			LastSyncNumber: 10,
			LastSyncHash:   common.Address{}.Hex(),
			UpdatedAt:      time.Now(),
		})
	})
	assert.NoError(t, err)

	res, err := blocksync.GetBlockSync(ctx, db, addr)
	assert.NoError(t, err)
	assert.Equal(t, addr, res.Address)
	assert.Equal(t, uint64(10), res.LastSyncNumber)
	assert.Equal(t, common.Address{}.Hex(), res.LastSyncHash)
}
