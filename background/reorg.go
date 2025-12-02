package background

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/eth"
	"evm_event_indexer/internal/storage"

	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"fmt"
	"log/slog"
	"time"
)

var _ Worker = (*ReorgConsumer)(nil)

type ReorgConsumer struct{}

func NewReorgConsumer() *ReorgConsumer {
	return &ReorgConsumer{}
}

func (r *ReorgConsumer) Run(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-reorgChan:
			// exceed retry limit, reset retry counter and skip the message
			if msg.Retry > config.Get().Retry {
				slog.Error("failed to handle reorg, exceed retry limit", slog.Any("retry", msg.Retry))
				continue
			}

			if err := r.reorgHandler(msg.LastSyncNumber, msg.ContractAddress); err != nil {
				slog.Error("failed to handle reorg", slog.Any("err", err))
				select {
				case <-ctx.Done():
					return nil
				default:
					// backoff
					time.Sleep(msg.Backoff)
					msg.Retry++
					msg.Backoff = min(msg.Backoff*2, config.Get().MaxBackoff)

					// requeue
					ReorgProducer(msg)
					continue
				}
			}
		}
	}
}

type reorgMsg struct {
	ContractAddress string
	LastSyncNumber  uint64
	Backoff         time.Duration
	Retry           int
}

var reorgChan = make(chan *reorgMsg, 1000)

func ReorgProducer(msg *reorgMsg) {

	for range config.Get().Retry {
		select {
		case reorgChan <- msg:
			return
		default:
			slog.Error("reorg channel is full, waiting for retry", slog.Any("msg", msg))
			time.Sleep(msg.Backoff)
			msg.Backoff = min(msg.Backoff*2, config.Get().MaxBackoff)
		}
	}

	slog.Error("reorg channel is full, msg dropped", slog.Any("msg", msg))
}

func (r *ReorgConsumer) reorgHandler(lastSyncNumber uint64, address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	// lock by address

	client, err := eth.NewClient(ctx, config.Get().EthRpcHTTP)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	checkpoint := lastSyncNumber
	found := false
	reorgHash := ""
	// start from page 1
	page := int32(1)
	// batch size
	size := config.Get().ReorgWindow
	for {

		// get batch logs to find the rollback checkpoint
		logs, err := eventlog.GetLogs(ctx, &eventlog.GetLogParam{
			Address:        address,
			OrderBy:        2,
			Desc:           true,
			BlockNumberLTE: checkpoint,
			Pagination: &model.Pagination{
				Page: page,
				Size: size,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}

		for _, log := range logs {
			// get the block on chain
			header, err := client.GetHeaderByNumber(log.BlockNumber)
			if err != nil {
				return fmt.Errorf("failed to get header: %w", err)
			}

			// if block hash matches, means we have found the checkpoint
			if header.Hash().String() == log.BlockHash {
				found = true
				reorgHash = header.Hash().String()
				break
			}

			// move checkpoint
			checkpoint = log.BlockNumber
		}

		// if checkpoint found, break
		// if less than size, means we are current at the last page, also break
		if int32(len(logs)) < size || found {
			break
		}

		// otherwise, move on to next page
		page++
	}

	now := time.Now()

	err = utils.NewTx(storage.GetMysql(config.Get().MySQL.EventDBS.DBName)).Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			return eventlog.TxDeleteLog(ctx, tx, address, checkpoint, lastSyncNumber)
		},
		func(ctx context.Context, tx *sql.Tx) error {
			return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
				Address:        address,
				LastSyncNumber: checkpoint,
				LastSyncHash:   reorgHash,
				UpdatedAt:      now,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute reorg tx: %w", err)
	}

	return nil
}
