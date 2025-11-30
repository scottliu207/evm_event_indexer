package utils

import (
	"context"
	"database/sql"
	"log"
)

type (
	Tx struct {
		db *sql.DB
	}
	FN func(ctx context.Context, tx *sql.Tx) error
)

func NewTx(db *sql.DB) *Tx {
	return &Tx{
		db: db,
	}
}

func (t *Tx) Exec(ctx context.Context, txFNs ...FN) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("[SQL 回滾失敗] Err:%s", err.Error())
		}
	}()

	for _, fn := range txFNs {
		if err := fn(ctx, tx); err != nil {
			return err
		}

	}

	return tx.Commit()
}
