package blocksync

import (
	"context"
	"database/sql"
	"evm_event_indexer/service/model"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// Upsert block sync status into db
func TxUpsertBlock(ctx context.Context, tx *sql.Tx, blockSync *model.BlockSync) error {

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Insert(model.TableNameBlockSync).
		Columns(
			"chain_id",
			"address",
			"last_sync_number",
			"last_sync_hash",
			"updated_at",
		).
		Values(
			blockSync.ChainID,
			blockSync.Address,
			blockSync.LastSyncNumber,
			blockSync.LastSyncHash,
			blockSync.UpdatedAt,
		).
		Suffix(`
	ON DUPLICATE KEY UPDATE 
		last_sync_number = VALUES(last_sync_number), 
		last_sync_hash = VALUES(last_sync_hash), 
		updated_at = VALUES(updated_at) 
	`)
	_, err := qb.RunWith(tx).ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func GetBlockSync(ctx context.Context, db *sql.DB, chainID int64, address string) (res *model.BlockSync, err error) {

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select(
			"chain_id",
			"address",
			"last_sync_number",
			"last_sync_hash",
			"updated_at",
		).
		From(model.TableNameBlockSync).
		Where(
			sq.Eq{"chain_id": chainID},
			sq.Eq{"address": address},
		)

	rows, err := qb.RunWith(db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		res = new(model.BlockSync)
		if err := rows.Scan(
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
	if chainID == 0 {
		return nil, fmt.Errorf("chain id is required")
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("addresses can not be empty")
	}

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select(
			"chain_id",
			"address",
			"last_sync_number",
			"last_sync_hash",
			"updated_at",
		).
		From(model.TableNameBlockSync).
		Where(
			sq.Eq{"chain_id": chainID},
			sq.Eq{"address": addresses},
		)

	rows, err := qb.RunWith(db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res = make(map[string]*model.BlockSync)
	for rows.Next() {
		bs := new(model.BlockSync)
		if err := rows.Scan(
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
