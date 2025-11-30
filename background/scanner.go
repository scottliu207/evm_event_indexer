package background

import (
	"context"
	"database/sql"
	internalCnf "evm_event_indexer/internal/config"
	internalEth "evm_event_indexer/internal/eth"
	internalStorage "evm_event_indexer/internal/storage"
	"fmt"

	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ERC20_TRANSFER_EVENT = "Transfer(address,address,uint256)"
	ERC20_APPROVAL_EVENT = "Approval(address,address,uint256)"
	size                 = uint64(100)
)

// LogScanner runs a periodic log sync for a specific contract address.
// It automatically retries with exponential backoff when the scanner stops unexpectedly.
func LogScanner(address string) {
	ctx := context.Background()
	backoff := internalCnf.Get().Backoff
	for {
		err := func() error {
			client, err := internalEth.NewClient(ctx, internalCnf.Get().EthRpcHTTP)
			if err != nil {
				return fmt.Errorf("failed to create eth client: %w", err)
			}
			defer client.Close()

			return scan(ctx, client, address)
		}()

		if err != nil {
			slog.Error("scanner error occurred, waiting to retry",
				slog.Any("err", err),
				slog.String("address", address),
			)
		}

		backoff = min(backoff*2, internalCnf.Get().MaxBackoff)
		slog.Info("waiting to retry", slog.Duration("duration", backoff))
		time.Sleep(backoff)
	}
}

func scan(ctx context.Context, client *internalEth.Client, address string) error {

	topics := [][]common.Hash{
		// topics, filetering Transfer and Approval events
		{
			common.Hash(crypto.Keccak256([]byte(ERC20_TRANSFER_EVENT))),
			common.Hash(crypto.Keccak256([]byte(ERC20_APPROVAL_EVENT))),
		},
		// from
		// to
	}

	ticker := time.NewTicker(internalCnf.Get().LogScannerInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := syncLog(ctx, client, address, topics); err != nil {
			slog.Error("syncLog error occurred",
				slog.Any("err", err),
				slog.String("address", address),
			)
		}
	}

	return nil
}

func syncLog(ctx context.Context, client *internalEth.Client, address string, topics [][]common.Hash) error {

	now := time.Now()

	bc, err := blocksync.GetBlockSync(ctx, internalStorage.GetMysql(internalCnf.Get().MySQL.EventDBS.DBName), address)
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

	toBlock := min(syncBlock+size, latestBlock)
	if syncBlock > toBlock {
		slog.Info("no new blocks to scan",
			slog.Any("lastSyncNumber", bc.LastSyncNumber),
			slog.Any("latestBlock", latestBlock),
		)
		return nil
	}

	eventLogs, err := client.GetLogs(internalEth.GetLogsParams{
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

	slog.Info("event logs info",
		slog.Any("lastSyncNumber", syncBlock),
		slog.Any("toBlock", toBlock),
		slog.Any("total logs count", len(eventLogs)),
	)

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
			TxIndex:        v.TxIndex,
			LogIndex:       v.Index,
			TxHash:         v.TxHash.Hex(),
			Data:           v.Data,
			BlockTimestamp: time.Unix(int64(v.BlockTimestamp), 0),
			CreatedAt:      now,
		}

		slog.Info("event log info",
			slog.Any("log", v),
		)
	}

	if err := utils.NewTx(internalStorage.GetMysql(internalCnf.Get().MySQL.EventDBS.DBName)).Exec(ctx,
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
