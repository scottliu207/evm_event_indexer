package eventlog

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/model"
	"strings"
	"time"
)

// Upsert log into db
func TxUpsertLog(ctx context.Context, tx *sql.Tx, log ...*model.Log) error {
	var sql strings.Builder
	var params []any
	var placeholder = make([]string, len(log))
	sql.WriteString(" INSERT INTO `event_db`.`event_log`( ")
	sql.WriteString("	`chain_id`, ")
	sql.WriteString("	`address`, ")
	sql.WriteString("	`block_hash`, ")
	sql.WriteString("	`block_number`, ")
	sql.WriteString("	`topics`, ")
	sql.WriteString("	`tx_index`, ")
	sql.WriteString("	`log_index`, ")
	sql.WriteString("	`tx_hash`, ")
	sql.WriteString("	`data`, ")
	sql.WriteString("	`decoded_event`, ")
	sql.WriteString("	`block_timestamp`, ")
	sql.WriteString("	`created_at` ")
	sql.WriteString(" ) VALUES ")

	for i, v := range log {
		placeholder[i] = " (?,?,?,?,?,?,?,?,?,?,?,?) "
		params = append(params, v.ChainID)
		params = append(params, v.Address)
		params = append(params, v.BlockHash)
		params = append(params, v.BlockNumber)
		params = append(params, v.Topics)
		params = append(params, v.TxIndex)
		params = append(params, v.LogIndex)
		params = append(params, v.TxHash)
		params = append(params, v.Data)
		params = append(params, v.DecodedEvent)
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

type GetLogParam struct {
	ChainID        int64
	Address        string
	OrderBy        int8 // 1:block_timestamp 2:block_number
	StartTime      time.Time
	EndTime        time.Time
	BlockNumberLTE uint64
	BlockNumberGTE uint64
	TxHash         string
	Desc           bool
	Pagination     *model.Pagination
}

// retrieves event logs matching the filter criteria.
func GetLogs(ctx context.Context, db *sql.DB, filter *GetLogParam) ([]*model.Log, error) {
	var sql strings.Builder
	var wheres []string
	var params []any

	sql.WriteString(" SELECT ")
	sql.WriteString("   id,")
	sql.WriteString("   chain_id,")
	sql.WriteString("   address,")
	sql.WriteString("   block_hash,")
	sql.WriteString("   block_number,")
	sql.WriteString("   topics,")
	sql.WriteString("   tx_index,")
	sql.WriteString("   log_index,")
	sql.WriteString("   tx_hash,")
	sql.WriteString("   data,")
	sql.WriteString("   decoded_event,")
	sql.WriteString("   block_timestamp,")
	sql.WriteString("   created_at ")
	sql.WriteString(" FROM event_db.event_log ")
	sql.WriteString(" WHERE ")

	if filter.ChainID != 0 {
		wheres = append(wheres, " chain_id = ? ")
		params = append(params, filter.ChainID)
	}

	if filter.Address != "" {
		wheres = append(wheres, " address = ? ")
		params = append(params, filter.Address)
	}

	if filter.TxHash != "" {
		wheres = append(wheres, " tx_hash = ? ")
		params = append(params, filter.TxHash)
	}

	if !filter.StartTime.IsZero() {
		params = append(params, filter.StartTime.UTC())
		wheres = append(wheres, " block_timestamp >= ? ")
	}

	if !filter.EndTime.IsZero() {
		params = append(params, filter.EndTime.UTC())
		wheres = append(wheres, " block_timestamp <= ? ")
	}

	if filter.BlockNumberLTE > 0 {
		params = append(params, filter.BlockNumberLTE)
		wheres = append(wheres, " block_number <= ? ")
	}

	if filter.BlockNumberGTE > 0 {
		params = append(params, filter.BlockNumberGTE)
		wheres = append(wheres, " block_number >= ? ")
	}

	sql.WriteString(strings.Join(wheres, " AND "))

	sql.WriteString(" ORDER BY ")
	switch filter.OrderBy {
	case 1:
		sql.WriteString(" block_timestamp ")
	case 2:
		sql.WriteString(" block_number ")
	default:
		sql.WriteString(" id ")
	}

	if filter.Desc {
		sql.WriteString(" DESC ")
	} else {
		sql.WriteString(" ASC ")
	}

	sql.WriteString(" LIMIT ? OFFSET ?")

	params = append(params, filter.Pagination.Limit())
	params = append(params, filter.Pagination.Offset())

	rows, err := db.QueryContext(ctx, sql.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.Log
	for rows.Next() {
		log := new(model.Log)
		if err := rows.Scan(
			&log.ID,
			&log.ChainID,
			&log.Address,
			&log.BlockHash,
			&log.BlockNumber,
			&log.Topics,
			&log.TxIndex,
			&log.LogIndex,
			&log.TxHash,
			&log.Data,
			&log.DecodedEvent,
			&log.BlockTimestamp,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// gets the total count of event logs matching the filter criteria.
func GetTotal(ctx context.Context, db *sql.DB, filter *GetLogParam) (int64, error) {

	var sql strings.Builder
	var wheres []string
	var params []any

	sql.WriteString(" SELECT ")
	sql.WriteString("   COUNT(*) ")
	sql.WriteString(" FROM event_db.event_log ")
	sql.WriteString(" WHERE ")

	if filter.ChainID != 0 {
		wheres = append(wheres, " chain_id = ? ")
		params = append(params, filter.ChainID)
	}

	if filter.Address != "" {
		wheres = append(wheres, " address = ? ")
		params = append(params, filter.Address)
	}

	if filter.TxHash != "" {
		wheres = append(wheres, " tx_hash = ? ")
		params = append(params, filter.TxHash)
	}

	if !filter.StartTime.IsZero() {
		params = append(params, filter.StartTime.UTC())
		wheres = append(wheres, " block_timestamp >= ? ")
	}

	if !filter.EndTime.IsZero() {
		params = append(params, filter.EndTime.UTC())
		wheres = append(wheres, " block_timestamp <= ? ")
	}

	if filter.BlockNumberLTE > 0 {
		params = append(params, filter.BlockNumberLTE)
		wheres = append(wheres, " block_number <= ? ")
	}

	if filter.BlockNumberGTE > 0 {
		params = append(params, filter.BlockNumberGTE)
		wheres = append(wheres, " block_number >= ? ")
	}

	sql.WriteString(strings.Join(wheres, " AND "))

	var total int64
	err := db.QueryRowContext(ctx, sql.String(), params...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func TxDeleteLog(ctx context.Context, tx *sql.Tx, address string, fromBN uint64, toBN uint64) error {
	const sql = `
		DELETE FROM event_db.event_log
		WHERE 
		  address = ?
		  AND block_number >= ?
		  AND block_number <= ?
	`
	var params []any

	params = append(params, address)
	params = append(params, fromBN)
	params = append(params, toBN)

	_, err := tx.ExecContext(ctx, sql, params...)
	if err != nil {
		return err
	}

	return nil
}
