package background

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/decoder"
	"evm_event_indexer/internal/enum"
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

var _ Worker = (*Scanner)(nil)

type Scanner struct {
	Address   string
	Topics    [][]common.Hash
	BatchSize int32
}

func NewScanner(address string, topics [][]common.Hash, batchSize int32) *Scanner {
	return &Scanner{
		Address:   address,
		Topics:    topics,
		BatchSize: batchSize,
	}
}

// Runs a periodic log sync for a specific contract address.
func (s *Scanner) Run(ctx context.Context) error {
	client, err := eth.NewClient(ctx, config.Get().EthRpcHTTP)
	if err != nil {
		return fmt.Errorf("failed to create eth client: %w", err)
	}

	defer client.Close()

	if err := s.scan(ctx, client); err != nil {
		return fmt.Errorf("scanner error: %w, address: %s", err, s.Address)
	}

	return nil
}

func (s *Scanner) scan(ctx context.Context, client *eth.Client) error {

	ticker := time.NewTicker(config.Get().LogScannerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.syncLog(ctx, client); err != nil {
				slog.Error("syncLog error",
					slog.Any("err", err),
					slog.String("address", s.Address),
				)
			}
		}
	}
}

func (s *Scanner) syncLog(ctx context.Context, client *eth.Client) error {

	now := time.Now()

	bc, err := blocksync.GetBlockSync(ctx, storage.GetMysql(config.Get().MySQL.EventDBS.DBName), s.Address)
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
		return fmt.Errorf("get current block number error for address %s: %w", s.Address, err)
	}

	toBlock := min(syncBlock+uint64(s.BatchSize), latestBlock)
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
		Addresses: []common.Address{common.HexToAddress(s.Address)},
		Topics:    s.Topics,
	})
	if err != nil {
		return fmt.Errorf("get logs error for address %s from block %d to %d: %w", s.Address, syncBlock, toBlock, err)
	}

	header, err := client.GetHeaderByNumber(toBlock)
	if err != nil {
		return fmt.Errorf("get block header error for block %d: %w", toBlock, err)
	}

	newSyncNumber := toBlock
	newSyncHash := header.Hash().Hex()
	logs := make([]*model.Log, len(eventLogs))
	for i, v := range eventLogs {
		logs[i] = &model.Log{
			ChainType:      enum.CHOther,
			Address:        v.Address.Hex(),
			BlockHash:      v.BlockHash.Hex(),
			BlockNumber:    v.BlockNumber,
			Topics:         (*model.Topics)(&v.Topics),
			TxIndex:        int32(v.TxIndex),
			LogIndex:       int32(v.Index),
			TxHash:         v.TxHash.Hex(),
			Data:           v.Data,
			BlockTimestamp: time.Unix(int64(v.BlockTimestamp), 0),
			CreatedAt:      now,
		}

		// decode event, if decode failed, keep raw data only
		name, args, err := decoder.Provider.Decode(logs[i])
		if err != nil {
			slog.Error("decode event error",
				slog.Any("err", err),
				slog.Any("address", s.Address),
				slog.Any("blockNumber", v.BlockNumber),
				slog.Any("txHash", v.TxHash.Hex()),
				slog.Any("logIndex", v.Index),
			)
		} else {
			logs[i].DecodedEvent = &model.DecodedEvent{
				EventName: name,
				EventData: args,
			}
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
				Address:        s.Address,
				LastSyncNumber: newSyncNumber,
				LastSyncHash:   newSyncHash,
				UpdatedAt:      now,
			})
		},
	); err != nil {
		return fmt.Errorf("upsert log error for address %s: %w", s.Address, err)
	}

	return nil
}
