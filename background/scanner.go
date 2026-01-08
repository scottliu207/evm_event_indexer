package background

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/decoder"
	"evm_event_indexer/internal/eth"
	"evm_event_indexer/service"
	"fmt"

	"evm_event_indexer/service/model"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var _ Worker = (*Scanner)(nil)

type Scanner struct {
	rpcHTTP   string
	Address   string
	Topics    [][]common.Hash
	BatchSize int32
}

func NewScanner(rcpHttp string, address string, topics [][]common.Hash, batchSize int32) *Scanner {
	return &Scanner{
		rpcHTTP:   rcpHttp,
		Address:   address,
		Topics:    topics,
		BatchSize: batchSize,
	}
}

// Runs a periodic log sync for a specific contract address.
func (s *Scanner) Run(ctx context.Context) error {
	client, err := eth.NewClient(ctx, s.rpcHTTP)
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
					slog.Any("error", err),
					slog.String("address", s.Address),
				)
			}
		}
	}
}

func (s *Scanner) syncLog(ctx context.Context, client *eth.Client) error {

	now := time.Now()

	ctx, cancel := context.WithTimeout(ctx, config.Get().Timeout)
	defer cancel()

	bc, err := service.GetBlockSync(ctx, client.GetChainID().Int64(), s.Address)
	if err != nil {
		return fmt.Errorf("get block sync status error: %w", err)
	}

	if bc == nil {
		bc = new(model.BlockSync)
	}

	// default start from block 0
	syncBlock := uint64(config.Get().StartBlock)

	// if there is no sync block, start from 0
	if bc.LastSyncNumber > 0 {
		syncBlock = bc.LastSyncNumber + 1
	}

	latestBlock, err := client.GetBlockNumber()
	if err != nil {
		return fmt.Errorf("get current block number error for address %s: %w", s.Address, err)
	}

	toBlock := min(syncBlock+uint64(s.BatchSize), latestBlock)
	if syncBlock >= toBlock {
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
		topics := make([]string, 4)
		ti := 0
		for _, t := range v.Topics {
			topics[ti] = t.Hex()
			ti++
		}

		logs[i] = &model.Log{
			ChainID:        client.GetChainID().Int64(),
			Address:        v.Address.Hex(),
			BlockHash:      v.BlockHash.Hex(),
			BlockNumber:    v.BlockNumber,
			Topic0:         topics[0],
			Topic1:         topics[1],
			Topic2:         topics[2],
			Topic3:         topics[3],
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
				slog.Any("error", err),
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

	params := &service.UpsertLogParam{
		ChainID:        client.GetChainID().Int64(),
		Address:        s.Address,
		LastSyncNumber: newSyncNumber,
		LastSyncHash:   newSyncHash,
		Now:            now,
		Logs:           logs,
	}

	if err := service.UpsertLog(ctx, params); err != nil {
		return fmt.Errorf("upsert log error for address %s: %w", s.Address, err)
	}

	return nil
}
