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
	sql.WriteString("	`block_timestamp` ")
	sql.WriteString(" ) VALUES ")

	for i, v := range log {
		placeholder[i] = " (?,?,?,?,?,?,?,?,?) "
		params = append(params, v.Address)
		params = append(params, v.BlockHash)
		params = append(params, v.BlockNumber)
		params = append(params, v.Topics)
		params = append(params, v.TxIndex)
		params = append(params, v.LogIndex)
		params = append(params, v.TxHash)
		params = append(params, v.Data)
		params = append(params, v.BlockTimestamp)
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
