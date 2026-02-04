package background

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/eth"
	"evm_event_indexer/service"

	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/eventlog"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

var _ Worker = (*ReorgConsumer)(nil)

type reorgMsg struct {
	RpcHttp         string
	ContractAddress string
	Backoff         time.Duration
	Log             types.Log
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

			if err := r.reorgHandler(ctx, msg.RpcHttp, msg.Log, msg.ContractAddress); err != nil {
				slog.Error("failed to handle reorg", slog.Any("error", err))
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

// handles the reorg event
// 1. check target log is same as on chain
// 2. if not same, fallback to get window size logs to find the rollback checkpoint
// 3. if still not found, fallback to start_block
func (r *ReorgConsumer) reorgHandler(parentCtx context.Context, rpcHttp string, log types.Log, address string) error {
	ctx, cancel := context.WithTimeout(parentCtx, config.Get().Timeout)
	defer cancel()

	client, err := eth.NewClient(ctx, rpcHttp)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	checkpoint := log.BlockNumber
	reorgHash := log.BlockHash.Hex()
	window := uint64(config.Get().ReorgWindow * 2)

	// get the logs by block hash
	blockLog, err := service.GetLogs(ctx, &eventlog.GetLogParam{
		Address:   address,
		BlockHash: log.BlockHash.Hex(),
		Pagination: &model.Pagination{ // only need to get one log
			Page: 1,
			Size: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	var rollbackHeader *types.Header

	for _, log := range blockLog {
		// get the block on chain
		header, err := client.GetHeaderByNumber(log.BlockNumber)
		if err != nil {
			return fmt.Errorf("failed to get header: %w", err)
		}

		// if block hash matches, means we have found the checkpoint
		if header.Hash().String() == log.BlockHash {
			rollbackHeader = header
			reorgHash = header.Hash().Hex()
			break
		}
	}

	// if rollback header not found, fallback to get batch logs to find the rollback checkpoint
	if rollbackHeader == nil {
		logs, err := service.GetLogs(ctx, &eventlog.GetLogParam{
			Address:        address,
			OrderBy:        2,
			Desc:           true, // from latest to oldest
			BlockNumberLTE: checkpoint,
			Pagination: &model.Pagination{
				Page: 1,
				Size: window,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}

		// iterate the logs to find the checkpoint
		for _, log := range logs {
			// get the block on chain
			header, err := client.GetHeaderByNumber(log.BlockNumber)
			if err != nil {
				return fmt.Errorf("failed to get header: %w", err)
			}

			// if block hash matches, means we have found the checkpoint
			if header.Hash().String() == log.BlockHash {
				rollbackHeader = header
				reorgHash = header.Hash().Hex()
				break
			}

			// otherwise, move checkpoint
			checkpoint = log.BlockNumber
		}
	}

	// if rollback header found in the batch logs, use it as checkpoint
	if rollbackHeader != nil {
		checkpoint = rollbackHeader.Number.Uint64()
		reorgHash = rollbackHeader.Hash().Hex()
	} else {
		// if rollback header not found in the batch logs, means reorg falls outside the window, fallback to start_block
		checkpoint = config.Get().StartBlock
		slog.Debug("reorg checkpoint not found within windowlimit, fallback to block start_block",
			slog.Any("checkpoint", checkpoint),
			slog.Any("window", window),
			slog.Any("start_block", config.Get().StartBlock),
		)

		// get the start_block header from chain
		header, err := client.GetHeaderByNumber(checkpoint)
		if err != nil {
			return fmt.Errorf("failed to get header: %w", err)
		}
		reorgHash = header.Hash().Hex()
	}

	params := &service.ReorgLogParam{
		ChainID:    client.GetChainID().Int64(),
		Address:    address,
		Checkpoint: checkpoint,
		ReorgHash:  reorgHash,
		Now:        time.Now(),
	}

	if err := service.ReorgLog(ctx, params); err != nil {
		return fmt.Errorf("failed to execute reorg tx: %w", err)
	}

	return nil
}
