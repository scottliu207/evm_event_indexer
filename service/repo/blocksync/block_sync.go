package blocksync

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/model"
	"fmt"
	"strings"
)

// Upsert block sync status into db
func TxUpsertBlock(ctx context.Context, tx *sql.Tx, blockSync *model.BlockSync) error {
	const sql = `
	INSERT INTO event_db.block_sync(
		chain_id, 
		address, 
		last_sync_number, 
		last_sync_hash, 
		updated_at 
	) VALUES (?,?,?,?,?) 
	ON DUPLICATE KEY UPDATE 
		last_sync_number = VALUES(last_sync_number), 
		last_sync_hash = VALUES(last_sync_hash), 
		updated_at = VALUES(updated_at) 
	`
	var params []any

	params = append(params, blockSync.ChainID)
	params = append(params, blockSync.Address)
	params = append(params, blockSync.LastSyncNumber)
	params = append(params, blockSync.LastSyncHash)
	params = append(params, blockSync.UpdatedAt)

	_, err := tx.ExecContext(ctx, sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func GetBlockSync(ctx context.Context, db *sql.DB, chainID int64, address string) (res *model.BlockSync, err error) {
	var sql strings.Builder
	sql.WriteString(" SELECT ")
	sql.WriteString("  `chain_id`, ")
	sql.WriteString("  `address`, ")
	sql.WriteString("  `last_sync_number`, ")
	sql.WriteString("  `last_sync_hash`, ")
	sql.WriteString("  `updated_at` ")
	sql.WriteString(" FROM `event_db`.`block_sync` ")
	sql.WriteString(" WHERE ")
	sql.WriteString("  `chain_id` = ? ")
	sql.WriteString("  AND `address` = ? ")

	row, err := db.QueryContext(ctx, sql.String(), chainID, address)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		res = new(model.BlockSync)
		if err := row.Scan(
			&res.ChainID,
			&res.Address,
			&res.LastSyncNumber,
			&res.LastSyncHash,
			&res.UpdatedAt,
		); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func GetBlockSyncMap(ctx context.Context, db *sql.DB, chainID int64, addresses []string) (res map[string]*model.BlockSync, err error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("addresses can not be empty")
	}

	var sql strings.Builder
	var params []any
	sql.WriteString(" SELECT ")
	sql.WriteString("  `chain_id`, ")
	sql.WriteString("  `address`, ")
	sql.WriteString("  `last_sync_number`, ")
	sql.WriteString("  `last_sync_hash`, ")
	sql.WriteString("  `updated_at` ")
	sql.WriteString(" FROM `event_db`.`block_sync` ")
	sql.WriteString(" WHERE ")
	sql.WriteString("  `chain_id` = ? ")
	params = append(params, chainID)
	sql.WriteString("  AND `address` IN ( ?")
	for i, v := range addresses {
		if i > 0 {
			sql.WriteString(", ?")
		}
		params = append(params, v)
	}
	sql.WriteString(")")

	row, err := db.QueryContext(ctx, sql.String(), params...)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	res = make(map[string]*model.BlockSync)
	for row.Next() {
		bs := new(model.BlockSync)
		if err := row.Scan(
			&bs.ChainID,
			&bs.Address,
			&bs.LastSyncNumber,
			&bs.LastSyncHash,
			&bs.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res[bs.Address] = bs
	}

	return res, nil
}
