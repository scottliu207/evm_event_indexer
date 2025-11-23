package background

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/db"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

const TIMEOUT = 30 * time.Second

var reorgChan = make(chan uint64, 1000)
var mu sync.Mutex

func ReorgConsumer() {

	for newHeadNumber := range reorgChan {

		if err := reorgHandler(newHeadNumber); err != nil {
			slog.Error("failed to handle reorg", slog.Any("err", err))
			reorgChan <- newHeadNumber
			continue
		}

	}
}

func ReorgProducer(newblockNumber uint64) {
	reorgChan <- newblockNumber
}

func reorgHandler(newHeadNumber uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	mu.Lock()
	defer mu.Unlock()

	now := time.Now()

	// get the last block number
	blockSync, err := blocksync.GetBlockSync(ctx, db.GetMysql(db.EVENT_DB), os.Getenv("CONTRACT_ADDRESS"))
	if err != nil {
		return fmt.Errorf("failed to get last sync data: %w", err)
	}

	if blockSync == nil {
		blockSync = new(model.BlockSync)
	}

	err = utils.NewTx(db.GetMysql(db.EVENT_DB)).Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			return eventlog.TxDeleteLog(ctx, tx, os.Getenv("CONTRACT_ADDRESS"), newHeadNumber)
		},
		func(ctx context.Context, tx *sql.Tx) error {
			rollbackNumber := uint64(0)
			if newHeadNumber > 0 {
				rollbackNumber = newHeadNumber - 1
			}
			return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
				Address:        os.Getenv("CONTRACT_ADDRESS"),
				LastSyncNumber: rollbackNumber,
				LastSyncHash:   "",
				UpdatedAt:      now,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute reorg tx: %w", err)
	}

	return nil
}
