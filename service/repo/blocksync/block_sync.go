package blocksync

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/model"
	"strings"
)

// Upsert block sync status into db
func TxUpsertBlock(ctx context.Context, tx *sql.Tx, blockSync *model.BlockSync) error {

	var sql strings.Builder
	var params []any

	sql.WriteString(" INSERT INTO `event_db`.`block_sync`(")
	sql.WriteString("	`address`, ")
	sql.WriteString("	`last_sync_number`, ")
	sql.WriteString("	`last_sync_timestamp`, ")
	sql.WriteString("	`last_finalized_number`, ")
	sql.WriteString("	`last_finalized_timestamp`, ")
	sql.WriteString("	`updated_at` ")
	sql.WriteString(" ) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE ")
	sql.WriteString("	`last_sync_number` = VALUES(`last_sync_number`), ")
	sql.WriteString("	`last_sync_timestamp` = VALUES(`last_sync_timestamp`), ")
	sql.WriteString("	`last_finalized_number` = VALUES(`last_finalized_number`), ")
	sql.WriteString("	`last_finalized_timestamp` = VALUES(`last_finalized_timestamp`), ")
	sql.WriteString("	`updated_at` = VALUES(`updated_at`) ")

	params = append(params, blockSync.Address)
	params = append(params, blockSync.LastSyncNumber)
	params = append(params, blockSync.LastSyncTimestamp)
	params = append(params, blockSync.LastFinalizedNumber)
	params = append(params, blockSync.LastFinalizedTimestamp)
	params = append(params, blockSync.UpdatedAt)

	_, err := tx.ExecContext(ctx, sql.String(), params...)
	if err != nil {
		return err
	}

	return nil
}

func GetBlockSync(ctx context.Context, db *sql.DB, address string) (res *model.BlockSync, err error) {
	var sql strings.Builder
	sql.WriteString(" SELECT ")
	sql.WriteString("  `address`, ")
	sql.WriteString("  `last_sync_number`, ")
	sql.WriteString("  `last_sync_timestamp`, ")
	sql.WriteString("  `last_finalized_number`, ")
	sql.WriteString("  `last_finalized_timestamp`, ")
	sql.WriteString("  `updated_at` ")
	sql.WriteString(" FROM `event_db`.`block_sync` ")
	sql.WriteString(" WHERE ")
	sql.WriteString("  `address` = ? ")

	row, err := db.QueryContext(ctx, sql.String(), address)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		res = new(model.BlockSync)
		if err := row.Scan(
			&res.Address,
			&res.LastSyncNumber,
			&res.LastSyncTimestamp,
			&res.LastFinalizedNumber,
			&res.LastFinalizedTimestamp,
			&res.UpdatedAt,
		); err != nil {
			return nil, err
		}
	}

	return res, nil
}
