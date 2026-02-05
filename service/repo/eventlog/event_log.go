package eventlog

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/model"
	"time"

	sq "github.com/Masterminds/squirrel"
)

func TxInsertLog(ctx context.Context, tx *sql.Tx, log ...*model.Log) error {
	if len(log) == 0 {
		return nil
	}

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Insert(model.TableNameEventLog).
		Columns(
			"chain_id",
			"address",
			"block_hash",
			"block_number",
			"topic_0",
			"topic_1",
			"topic_2",
			"topic_3",
			"tx_index",
			"log_index",
			"tx_hash",
			"data",
			"decoded_event",
			"block_timestamp",
			"created_at",
		)

	for _, v := range log {
		qb = qb.Values(
			v.ChainID,
			v.Address,
			v.BlockHash,
			v.BlockNumber,
			v.Topic0,
			v.Topic1,
			v.Topic2,
			v.Topic3,
			v.TxIndex,
			v.LogIndex,
			v.TxHash,
			v.Data,
			v.DecodedEvent,
			v.BlockTimestamp,
			v.CreatedAt,
		)
	}

	_, err := qb.RunWith(tx).ExecContext(ctx)
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
	BlockHash      string
	Topic0         string
	Topic1         string
	Topic2         string
	Topic3         string
	Desc           bool
	Pagination     *model.Pagination
}

func (p GetLogParam) ToWhere() sq.And {
	var conds sq.And
	if p.ChainID != 0 {
		conds = append(conds, sq.Eq{"chain_id": p.ChainID})
	}

	if p.Address != "" {
		conds = append(conds, sq.Eq{"address": p.Address})
	}

	if p.TxHash != "" {
		conds = append(conds, sq.Eq{"tx_hash": p.TxHash})
	}

	if p.Topic0 != "" {
		conds = append(conds, sq.Eq{"topic_0": p.Topic0})
	}

	if p.Topic1 != "" {
		conds = append(conds, sq.Eq{"topic_1": p.Topic1})
	}

	if p.Topic2 != "" {
		conds = append(conds, sq.Eq{"topic_2": p.Topic2})
	}

	if p.Topic3 != "" {
		conds = append(conds, sq.Eq{"topic_3": p.Topic3})
	}

	if !p.StartTime.IsZero() {
		conds = append(conds, sq.GtOrEq{"block_timestamp": p.StartTime.UTC()})
	}

	if !p.EndTime.IsZero() {
		conds = append(conds, sq.LtOrEq{"block_timestamp": p.EndTime.UTC()})
	}

	if p.BlockNumberLTE > 0 {
		conds = append(conds, sq.LtOrEq{"block_number": p.BlockNumberLTE})
	}

	if p.BlockNumberGTE > 0 {
		conds = append(conds, sq.GtOrEq{"block_number": p.BlockNumberGTE})
	}

	if p.BlockHash != "" {
		conds = append(conds, sq.Eq{"block_hash": p.BlockHash})
	}

	return conds
}

func (p GetLogParam) ToOrderBy() string {
	var orderby = ""

	switch p.OrderBy {
	case 1:
		orderby = "block_timestamp"
	case 2:
		orderby = "block_number"
	default:
		orderby = "id"
	}

	if p.Desc {
		orderby += " DESC"
	} else {
		orderby += " ASC"
	}
	return orderby
}

func GetLogs(ctx context.Context, db *sql.DB, filter *GetLogParam) ([]*model.Log, error) {

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select(
			"id",
			"chain_id",
			"address",
			"block_hash",
			"block_number",
			"topic_0",
			"topic_1",
			"topic_2",
			"topic_3",
			"tx_index",
			"log_index",
			"tx_hash",
			"data",
			"decoded_event",
			"block_timestamp",
			"created_at",
		).
		From(model.TableNameEventLog).
		Where(filter.ToWhere()).
		OrderBy(filter.ToOrderBy()).
		Limit(filter.Pagination.Limit()).
		Offset(filter.Pagination.Offset())

	rows, err := qb.RunWith(db).QueryContext(ctx)
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
			&log.Topic0,
			&log.Topic1,
			&log.Topic2,
			&log.Topic3,
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
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select("COUNT(*)").
		From(model.TableNameEventLog).
		Where(filter.ToWhere())

	var total int64
	if err := qb.RunWith(db).QueryRowContext(ctx).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

// deletes event logs after a given block number
func TxDeleteLog(ctx context.Context, tx *sql.Tx, address string, fromBN uint64) error {

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Delete(model.TableNameEventLog).
		Where(
			sq.Eq{"address": address},
			sq.Gt{"block_number": fromBN},
		)

	_, err := qb.RunWith(tx).ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
