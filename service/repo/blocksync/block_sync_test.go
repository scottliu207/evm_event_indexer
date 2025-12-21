package blocksync_test

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/internal/testutil"
	"evm_event_indexer/service/model"
	"fmt"
	"os"

	"evm_event_indexer/service/repo/blocksync"

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

func Test_TxUpsertBlock(t *testing.T) {

	db, err := storage.GetMySQL(config.EventDBM)
	if err != nil {
		t.Fatalf("failed to get mysql: %s\n", err)
	}

	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	err = utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
			ChainID:        31337,
			Address:        addr,
			LastSyncNumber: 10,
			LastSyncHash:   common.Address{}.Hex(),
			UpdatedAt:      time.Now(),
		})
	})
	assert.NoError(t, err)

	res, err := blocksync.GetBlockSync(ctx, db, 31337, addr)
	assert.NoError(t, err)
	assert.Equal(t, addr, res.Address)
	assert.Equal(t, uint64(10), res.LastSyncNumber)
	assert.Equal(t, common.Address{}.Hex(), res.LastSyncHash)
}
