package background

import (
	"context"
	"database/sql"
	internalCnf "evm_event_indexer/internal/config"
	internalEth "evm_event_indexer/internal/eth"
	internalStorage "evm_event_indexer/internal/storage"

	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"fmt"
	"log/slog"
	"time"
)

type reorgMsg struct {
	RollbackNumber uint64
	Backoff        time.Duration
	Retry          int
}

var reorgChan = make(chan *reorgMsg, 1000)

func ReorgConsumer() {

	for msg := range reorgChan {
		// exceed retry limit, reset retry counter and skip the message
		if msg.Retry > internalCnf.Get().Retry {
			slog.Error("failed to handle reorg, exceed retry limit", slog.Any("retry", msg.Retry))
			continue
		}

		if err := reorgHandler(msg.RollbackNumber); err != nil {
			slog.Error("failed to handle reorg", slog.Any("err", err))

			// backoff
			time.Sleep(msg.Backoff)
			msg.Retry++
			msg.Backoff = min(msg.Backoff*2, internalCnf.Get().MaxBackoff)

			// requeue
			ReorgProducer(msg)
			continue
		}
	}
}

func ReorgProducer(msg *reorgMsg) {

	for range internalCnf.Get().Retry {
		select {
		case reorgChan <- msg:
			return
		default:
			slog.Error("reorg channel is full, waiting for retry", slog.Any("msg", msg))
			time.Sleep(msg.Backoff)
			msg.Backoff = min(msg.Backoff*2, internalCnf.Get().MaxBackoff)
		}
	}

	slog.Error("reorg channel is full, msg dropped", slog.Any("msg", msg))
}

func reorgHandler(rollbackNumber uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), internalCnf.Get().Timeout)
	defer cancel()

	// lock by address

	now := time.Now()

	// get the last block number
	blockSync, err := blocksync.GetBlockSync(ctx, internalStorage.GetMysql(internalCnf.Get().EventDB), internalCnf.Get().ContractAddress)
	if err != nil {
		return fmt.Errorf("failed to get last sync data: %w", err)
	}

	if blockSync == nil {
		blockSync = new(model.BlockSync)
	}

	prev := uint64(0)
	if rollbackNumber > 0 {
		prev = rollbackNumber - 1
	}

	client, err := internalEth.NewClient(ctx, internalCnf.Get().EthRpcHTTP)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// get the previous block header
	header, err := client.HeaderByNumber(prev)
	if err != nil {
		return fmt.Errorf("failed to get header: %w", err)
	}

	rollbackHash := header.Hash().String()

	err = utils.NewTx(internalStorage.GetMysql(internalCnf.Get().EventDB)).Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			return eventlog.TxDeleteLog(ctx, tx, internalCnf.Get().ContractAddress, rollbackNumber)
		},
		func(ctx context.Context, tx *sql.Tx) error {

			return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
				Address:        internalCnf.Get().ContractAddress,
				LastSyncNumber: prev,
				LastSyncHash:   rollbackHash,
				UpdatedAt:      now,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute reorg tx: %w", err)
	}

	return nil
}
