package admin

import (
	"context"
	"database/sql"

	"evm_event_indexer/internal/enum"
	"evm_event_indexer/service/model"

	sq "github.com/Masterminds/squirrel"
)

func TxInsertAdmin(ctx context.Context, tx *sql.Tx, admin *model.Admin) (int64, error) {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Insert(model.TableNameAdmin).
		Columns(
			"account",
			"status",
			"password",
			"auth_meta",
			"created_at",
		).
		Values(
			admin.Account,
			admin.Status,
			admin.Password,
			admin.AuthMeta,
			admin.CreatedAt,
		)

	result, err := qb.RunWith(tx).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func TxUpdateAdmin(ctx context.Context, tx *sql.Tx, admin *model.Admin) error {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Update(model.TableNameAdmin).
		Where(sq.Eq{"id": admin.ID})

	if admin.Account != "" {
		qb = qb.Set("account", admin.Account)
	}
	if admin.Status != 0 {
		qb = qb.Set("status", admin.Status)
	}
	if admin.Password != "" {
		qb = qb.Set("password", admin.Password)
	}
	if admin.AuthMeta != nil {
		qb = qb.Set("auth_meta", admin.AuthMeta)
	}

	_, err := qb.RunWith(tx).ExecContext(ctx)
	return err
}

func TxDeleteAdmin(ctx context.Context, tx *sql.Tx, adminID int64) error {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Update(model.TableNameAdmin).
		Set("status", enum.UserStatusDisabled).
		Where(sq.Eq{"id": adminID})

	_, err := qb.RunWith(tx).ExecContext(ctx)
	return err
}

type GetAdminFilter struct {
	IDs        []int64
	Accounts   []string
	Status     enum.UserStatus
	Pagination *model.Pagination
}

func (p GetAdminFilter) ToWhere() sq.And {
	var conds sq.And
	if len(p.IDs) > 0 {
		conds = append(conds, sq.Eq{"id": p.IDs})
	}
	if len(p.Accounts) > 0 {
		conds = append(conds, sq.Eq{"account": p.Accounts})
	}
	if p.Status != 0 {
		conds = append(conds, sq.Eq{"status": p.Status})
	}
	return conds
}

func (p GetAdminFilter) ToOrderBy() string {
	return "id"
}

func GetAdminTotal(ctx context.Context, db *sql.DB, filter *GetAdminFilter) (int64, error) {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select("COUNT(*)").
		From(model.TableNameAdmin).
		Where(filter.ToWhere())

	var total int64
	err := qb.RunWith(db).QueryRowContext(ctx).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func GetAdmins(ctx context.Context, db *sql.DB, filter *GetAdminFilter) ([]*model.Admin, error) {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Select(
			"id",
			"account",
			"status",
			"password",
			"auth_meta",
			"created_at",
			"updated_at",
		).
		From(model.TableNameAdmin).
		Where(filter.ToWhere()).
		OrderBy(filter.ToOrderBy())

	if filter != nil && filter.Pagination != nil {
		qb = qb.Offset(filter.Pagination.Offset()).Limit(filter.Pagination.Limit())
	}

	rows, err := qb.RunWith(db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*model.Admin, 0)
	for rows.Next() {
		admin := new(model.Admin)
		if err := rows.Scan(
			&admin.ID,
			&admin.Account,
			&admin.Status,
			&admin.Password,
			&admin.AuthMeta,
			&admin.CreatedAt,
			&admin.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, admin)
	}

	return res, nil
}
