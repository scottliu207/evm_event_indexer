package service

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"fmt"
	"time"
)

type UpsertLogParam struct {
	ChainID        int64
	Address        string
	LastSyncNumber uint64
	LastSyncHash   string
	Now            time.Time
	Logs           []*model.Log
}

// UpsertLog upserts event logs and block sync info into database.
func UpsertLog(ctx context.Context, params *UpsertLogParam) error {
	if params == nil {
		return fmt.Errorf("params is nil")
	}
	if len(params.Logs) == 0 {
		return fmt.Errorf("logs is nil")
	}
	if params.ChainID == 0 {
		return fmt.Errorf("chain id is 0")
	}
	if params.Address == "" {
		return fmt.Errorf("address is empty")
	}
	if params.LastSyncNumber == 0 {
		return fmt.Errorf("last sync number is 0")
	}
	if params.Now.IsZero() {
		return fmt.Errorf("now is zero")
	}
	if params.LastSyncHash == "" {
		return fmt.Errorf("last sync hash is empty")
	}

	db, err := storage.GetMySQL(config.EventDBM)
	if err != nil {
		return fmt.Errorf("failed to get mysql: %w", err)
	}

	if err := utils.NewTx(db).Exec(ctx,
		// lock the block sync record for update
		func(ctx context.Context, tx *sql.Tx) error {
			return blocksync.TxSelectForUpdateBlockSync(ctx, tx, params.ChainID, params.Address)
		},
		// delete the logs after the last sync number
		func(ctx context.Context, tx *sql.Tx) error {
			return eventlog.TxDeleteLog(ctx, tx, params.Address, params.LastSyncNumber)
		},
		// upsert the logs
		func(ctx context.Context, tx *sql.Tx) error {
			if len(params.Logs) == 0 {
				return nil
			}
			return eventlog.TxInsertLog(ctx, tx, params.Logs...)
		},
		// upsert the block sync record
		func(ctx context.Context, tx *sql.Tx) error {
			return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
				ChainID:        params.ChainID,
				Address:        params.Address,
				LastSyncNumber: params.LastSyncNumber,
				LastSyncHash:   params.LastSyncHash,
				UpdatedAt:      params.Now,
			})
		},
	); err != nil {
		return fmt.Errorf("upsert log error for address %s: %w", params.Address, err)
	}
	return nil
}

type ReorgLogParam struct {
	ChainID        int64
	Address        string
	Checkpoint     uint64
	LastSyncNumber uint64
	ReorgHash      string
	Now            time.Time
}

// ReorgLog deletes event logs in a given range and updates block sync info into database.
func ReorgLog(ctx context.Context, params *ReorgLogParam) error {

	if params == nil {
		return fmt.Errorf("params is nil")
	}
	if params.ChainID == 0 {
		return fmt.Errorf("chain id is 0")
	}
	if params.Address == "" {
		return fmt.Errorf("address is empty")
	}
	if params.Checkpoint == 0 {
		return fmt.Errorf("checkpoint is 0")
	}
	if params.Now.IsZero() {
		return fmt.Errorf("now is zero")
	}

	db, err := storage.GetMySQL(config.EventDBM)
	if err != nil {
		return fmt.Errorf("failed to get mysql: %w", err)
	}

	err = utils.NewTx(db).Exec(ctx,
		// lock the block sync record for update
		func(ctx context.Context, tx *sql.Tx) error {
			return blocksync.TxSelectForUpdateBlockSync(ctx, tx, params.ChainID, params.Address)
		},
		func(ctx context.Context, tx *sql.Tx) error {
			return eventlog.TxDeleteLog(ctx, tx, params.Address, params.Checkpoint)
		},
		func(ctx context.Context, tx *sql.Tx) error {
			return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
				ChainID:        params.ChainID,
				Address:        params.Address,
				LastSyncNumber: params.Checkpoint,
				LastSyncHash:   params.ReorgHash,
				UpdatedAt:      params.Now,
			})
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute reorg tx: %w", err)
	}

	return nil
}

// GetLogsWithTotal retrieves event logs and total counts matching the filter criteria.
func GetLogsWithTotal(ctx context.Context, filter *eventlog.GetLogParam) (logs []*model.Log, total int64, err error) {

	db, err := storage.GetMySQL(config.EventDBS)
	if err != nil {
		return nil, 0, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	total, err = eventlog.GetTotal(ctx, db, filter)
	if err != nil {
		return nil, 0, errors.ErrInternalServerError.Wrap(err, "failed to get total")
	}

	// no need to proceed if no data found
	if total == 0 {
		return nil, 0, nil
	}

	logs, err = eventlog.GetLogs(ctx, db, filter)
	if err != nil {
		return nil, 0, errors.ErrInternalServerError.Wrap(err, "failed to get logs")
	}

	return logs, total, nil
}

// GetLogs retrieves event logs matching the filter criteria.
func GetLogs(ctx context.Context, filter *eventlog.GetLogParam) (logs []*model.Log, err error) {
	db, err := storage.GetMySQL(config.EventDBS)
	if err != nil {
		return nil, fmt.Errorf("failed to get mysql: %w", err)
	}

	logs, err = eventlog.GetLogs(ctx, db, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return logs, nil
}

// GetBlockSync retrieves the block sync info for a given chain and address
func GetBlockSync(ctx context.Context, chainID int64, address string) (*model.BlockSync, error) {
	db, err := storage.GetMySQL(config.EventDBS)
	if err != nil {
		return nil, fmt.Errorf("failed to get mysql: %w", err)
	}

	return blocksync.GetBlockSync(ctx, db, chainID, address)
}

// GetBlockSyncMap retrieves the block sync info map for a given chain and addresses
func GetBlockSyncMap(ctx context.Context, chainID int64, addresses []string) (map[string]*model.BlockSync, error) {
	if chainID == 0 {
		return nil, fmt.Errorf("chain id is required")
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("addresses are required")
	}

	db, err := storage.GetMySQL(config.EventDBS)
	if err != nil {
		return nil, fmt.Errorf("failed to get mysql: %w", err)
	}

	return blocksync.GetBlockSyncMap(ctx, db, chainID, addresses)
}
