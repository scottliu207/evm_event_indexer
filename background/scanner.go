package background

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/eth"
	"evm_event_indexer/internal/storage"
	"fmt"

	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// LogScanner runs a periodic log sync for a specific contract address.
// It automatically retries with exponential backoff when the scanner stops unexpectedly.
func LogScanner(address string, topics [][]common.Hash, batchSize int32) {
	ctx := context.Background()

	for {
		err := func() error {
			client, err := eth.NewClient(ctx, config.Get().EthRpcHTTP)
			if err != nil {
				return fmt.Errorf("failed to create eth client: %w", err)
			}
			defer client.Close()

			return scan(ctx, client, address, topics, batchSize)
		}()

		if err != nil {
			slog.Error("scanner error occurred, waiting to retry",
				slog.Any("err", err),
				slog.Any("address", address),
				slog.Any("retry_delay", config.Get().MaxBackoff),
			)
			time.Sleep(config.Get().MaxBackoff)
		}
	}
}

func scan(ctx context.Context, client *eth.Client, address string, topics [][]common.Hash, batchSize int32) error {

	ticker := time.NewTicker(config.Get().LogScannerInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := syncLog(ctx, client, address, topics, batchSize); err != nil {
			slog.Error("syncLog error occurred",
				slog.Any("err", err),
				slog.String("address", address),
			)
		}
	}

	return nil
}

func syncLog(ctx context.Context, client *eth.Client, address string, topics [][]common.Hash, batchSize int32) error {

	now := time.Now()

	bc, err := blocksync.GetBlockSync(ctx, storage.GetMysql(config.Get().MySQL.EventDBS.DBName), address)
	if err != nil {
		return fmt.Errorf("get block sync status error: %w", err)
	}

	if bc == nil {
		bc = new(model.BlockSync)
	}

	// default start from block 0
	syncBlock := uint64(0)

	// if there is no sync block, start from 0
	if bc.LastSyncNumber > 0 {
		syncBlock = bc.LastSyncNumber + 1
	}

	latestBlock, err := client.GetBlockNumber()
	if err != nil {
		return fmt.Errorf("get current block number error for address %s: %w", address, err)
	}

	toBlock := min(syncBlock+uint64(batchSize), latestBlock)
	if syncBlock > toBlock {
		slog.Info("no new blocks to scan",
			slog.Any("lastSyncNumber", bc.LastSyncNumber),
			slog.Any("latestBlock", latestBlock),
		)
		return nil
	}

	eventLogs, err := client.GetLogs(eth.GetLogsParams{
		FromBlock: syncBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{common.HexToAddress(address)},
		Topics:    topics,
	})
	if err != nil {
		return fmt.Errorf("get logs error for address %s from block %d to %d: %w", address, syncBlock, toBlock, err)
	}

	header, err := client.GetHeaderByNumber(toBlock)
	if err != nil {
		return fmt.Errorf("get block header error for block %d: %w", toBlock, err)
	}

	newSyncNumber := toBlock
	newSyncHash := header.Hash().Hex()

	logs := make([]*model.Log, len(eventLogs))
	for i, v := range eventLogs {

		topics := make(model.Topics, len(v.Topics))
		for j, topic := range v.Topics {
			topics[j] = topic.Hex()
		}

		logs[i] = &model.Log{
			Address:        v.Address.Hex(),
			BlockHash:      v.BlockHash.Hex(),
			BlockNumber:    v.BlockNumber,
			Topics:         &topics,
			TxIndex:        int32(v.TxIndex),
			LogIndex:       int32(v.Index),
			TxHash:         v.TxHash.Hex(),
			Data:           v.Data,
			BlockTimestamp: time.Unix(int64(v.BlockTimestamp), 0),
			CreatedAt:      now,
		}
	}

	if err := utils.NewTx(storage.GetMysql(config.Get().MySQL.EventDBS.DBName)).Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			if len(logs) == 0 {
				return nil
			}
			return eventlog.TxUpsertLog(ctx, tx, logs...)
		},
		func(ctx context.Context, tx *sql.Tx) error {
			return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
				Address:        address,
				LastSyncNumber: newSyncNumber,
				LastSyncHash:   newSyncHash,
				UpdatedAt:      now,
			})
		},
	); err != nil {
		return fmt.Errorf("upsert log error for address %s: %w", address, err)
	}

	return nil
}
