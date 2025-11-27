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
	ContractAddress string
	RollbackNumber  uint64
	Backoff         time.Duration
	Retry           int
}

var reorgChan = make(chan *reorgMsg, 1000)

func ReorgConsumer() {

	for msg := range reorgChan {
		// exceed retry limit, reset retry counter and skip the message
		if msg.Retry > internalCnf.Get().Retry {
			slog.Error("failed to handle reorg, exceed retry limit", slog.Any("retry", msg.Retry))
			continue
		}

		if err := reorgHandler(msg.RollbackNumber, msg.ContractAddress); err != nil {
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

func reorgHandler(rollbackNumber uint64, address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), internalCnf.Get().Timeout)
	defer cancel()

	// lock by address

	client, err := internalEth.NewClient(ctx, internalCnf.Get().EthRpcHTTP)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	rollbackSyncNumber := uint64(0)
	if rollbackNumber > 0 {
		rollbackSyncNumber = rollbackNumber - 1
	}

	header, err := client.GetHeaderByNumber(rollbackSyncNumber)
	if err != nil {
		return fmt.Errorf("failed to get header: %w", err)
	}

	now := time.Now()

	err = utils.NewTx(internalStorage.GetMysql(internalCnf.Get().EventDB)).Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			return eventlog.TxDeleteLog(ctx, tx, address, rollbackNumber)
		},
		func(ctx context.Context, tx *sql.Tx) error {
			return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
				Address:        address,
				LastSyncNumber: rollbackSyncNumber,
				LastSyncHash:   header.Hash().String(),
				UpdatedAt:      now,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute reorg tx: %w", err)
	}

	return nil
}
