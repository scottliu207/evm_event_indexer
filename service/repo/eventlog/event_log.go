package eventlog

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/model"
	"strings"
)

// Upsert log into db
func TxUpsertLog(ctx context.Context, tx *sql.Tx, log ...*model.Log) error {
	var sql strings.Builder
	var params []any
	var placeholder = make([]string, len(log))
	sql.WriteString(" INSERT INTO `event_db`.`event_log`( ")
	sql.WriteString("	`address`, ")
	sql.WriteString("	`block_hash`, ")
	sql.WriteString("	`block_number`, ")
	sql.WriteString("	`topics`, ")
	sql.WriteString("	`tx_index`, ")
	sql.WriteString("	`log_index`, ")
	sql.WriteString("	`tx_hash`, ")
	sql.WriteString("	`data`, ")
	sql.WriteString("	`block_timestamp`, ")
	sql.WriteString("	`created_at` ")
	sql.WriteString(" ) VALUES ")

	for i, v := range log {
		placeholder[i] = " (?,?,?,?,?,?,?,?,?,?) "
		params = append(params, v.Address)
		params = append(params, v.BlockHash)
		params = append(params, v.BlockNumber)
		params = append(params, v.Topics)
		params = append(params, v.TxIndex)
		params = append(params, v.LogIndex)
		params = append(params, v.TxHash)
		params = append(params, v.Data)
		params = append(params, v.BlockTimestamp)
		params = append(params, v.CreatedAt)
	}

	sql.WriteString(strings.Join(placeholder, ","))

	sql.WriteString(" ON DUPLICATE KEY UPDATE")
	sql.WriteString("	`block_hash` = VALUES(`block_hash`), ")
	sql.WriteString("	`topics` = VALUES(`topics`), ")
	sql.WriteString("	`tx_hash` = VALUES(`tx_hash`), ")
	sql.WriteString("	`data` = VALUES(`data`), ")
	sql.WriteString("	`block_timestamp` = VALUES(`block_timestamp`) ")

	_, err := tx.ExecContext(ctx, sql.String(), params...)
	if err != nil {
		return err
	}

	return nil
}

func TxGetEventLog(ctx context.Context, tx *sql.Tx, address string, blockNumber uint64) ([]*model.Log, error) {
	var sql strings.Builder
	var params []any

	sql.WriteString(" SELECT ")
	sql.WriteString("   `address`, ")
	sql.WriteString("   `block_hash`, ")
	sql.WriteString("   `block_number`, ")
	sql.WriteString("   `topics`, ")
	sql.WriteString("   `tx_index`, ")
	sql.WriteString("   `log_index`, ")
	sql.WriteString("   `tx_hash`, ")
	sql.WriteString("   `data`, ")
	sql.WriteString("   `block_timestamp`, ")
	sql.WriteString("   `created_at` ")
	sql.WriteString(" FROM `event_db`.`event_log` ")
	sql.WriteString(" WHERE ")
	sql.WriteString("   `address` = ? ")
	sql.WriteString("   AND `block_number` >= ? ")
	sql.WriteString(" ORDER BY `block_number` ASC ")

	params = append(params, address)
	params = append(params, blockNumber)

	rows, err := tx.QueryContext(ctx, sql.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.Log
	for rows.Next() {
		log := new(model.Log)
		if err := rows.Scan(
			&log.Address,
			&log.BlockHash,
			&log.BlockNumber,
			&log.Topics,
			&log.TxIndex,
			&log.LogIndex,
			&log.TxHash,
			&log.Data,
			&log.BlockTimestamp,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func TxDeleteLog(ctx context.Context, tx *sql.Tx, address string, blockNumber uint64) error {
	var sql strings.Builder
	var params []any

	sql.WriteString(" DELETE FROM `event_db`.`event_log`  ")
	sql.WriteString(" WHERE ")
	sql.WriteString("	`address` = ? ")
	sql.WriteString("	AND `block_number` >= ? ")

	params = append(params, address)
	params = append(params, blockNumber)

	_, err := tx.ExecContext(ctx, sql.String(), params...)
	if err != nil {
		return err
	}

	return nil
}

func GetEventLog(ctx context.Context, db *sql.DB, id int64) (res *model.Log, err error) {
	var sql strings.Builder

	sql.WriteString(" SELECT ")
	sql.WriteString("  `address`, ")
	sql.WriteString("  `block_hash`, ")
	sql.WriteString("  `block_number`, ")
	sql.WriteString("  `topics`, ")
	sql.WriteString("  `tx_index`, ")
	sql.WriteString("  `log_index`, ")
	sql.WriteString("  `tx_hash`, ")
	sql.WriteString("  `data`, ")
	sql.WriteString("  `block_timestamp`, ")
	sql.WriteString("  `created_at` ")
	sql.WriteString(" FROM `event_db`.`event_log` ")
	sql.WriteString(" WHERE ")
	sql.WriteString("  `id` = ? ")

	row, err := db.QueryContext(ctx, sql.String(), id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		res = new(model.Log)
		if err := row.Scan(
			&res.Address,
			&res.BlockHash,
			&res.BlockNumber,
			&res.Topics,
			&res.TxIndex,
			&res.LogIndex,
			&res.TxHash,
			&res.Data,
			&res.BlockTimestamp,
			&res.CreatedAt,
		); err != nil {
			return nil, err
		}
	}

	return res, nil
}
