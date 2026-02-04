package user

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/service/model"

	sq "github.com/Masterminds/squirrel"
)

// Insert user info into db
func TxInsertUser(ctx context.Context, tx *sql.Tx, user *model.User) (int64, error) {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Insert(model.TableNameUser).
		Columns(
			"account",
			"status",
			"password",
			"auth_meta",
			"created_at",
		).
		Values(
			user.Account,
			user.Status,
			user.Password,
			user.AuthMeta,
			user.CreatedAt,
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

// Delete user by ID
func TxDeleteUser(ctx context.Context, tx *sql.Tx, userID int64) error {

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Delete(model.TableNameUser).
		Where(sq.Eq{"id": userID})

	_, err := qb.RunWith(tx).ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

type GetUserFilter struct {
	Accounts []string
	Status   enum.UserStatus
}

func (p GetUserFilter) ToWhere() sq.And {
	var conds sq.And
	if len(p.Accounts) > 0 {
		conds = append(conds, sq.Eq{"account": p.Accounts})
	}
	if p.Status != 0 {
		conds = append(conds, sq.Eq{"status": p.Status})
	}
	return conds
}

func (p GetUserFilter) ToOrderBy() string {
	return "id"
}

func GetUsers(ctx context.Context, db *sql.DB, filter *GetUserFilter) ([]*model.User, error) {

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
		From(model.TableNameUser).
		Where(filter.ToWhere()).
		OrderBy(filter.ToOrderBy())

	rows, err := qb.RunWith(db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*model.User, 0)
	for rows.Next() {
		user := new(model.User)
		if err := rows.Scan(
			&user.ID,
			&user.Account,
			&user.Status,
			&user.Password,
			&user.AuthMeta,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, user)
	}

	return res, nil
}
