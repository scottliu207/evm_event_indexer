package user

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/service/model"
	"strings"
)

// Insert user info into db
func TxInsertUser(ctx context.Context, tx *sql.Tx, user *model.User) (int64, error) {
	const sql = `
	INSERT INTO account_db.user (
		account,
		status,
		role,
		password,
		auth_meta,
		created_at
	) VALUES (?, ?, ?, ?, ?, ?)
`

	var params []any

	params = append(params, user.Account)
	params = append(params, user.Status)
	params = append(params, user.Role)
	params = append(params, user.Password)
	params = append(params, user.AuthMeta)
	params = append(params, user.CreatedAt)

	result, err := tx.ExecContext(ctx, sql, params...)
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
	sql := `
	DELETE FROM account_db.user 
	WHERE id = ?
`
	_, err := tx.ExecContext(ctx, sql, userID)
	if err != nil {
		return err
	}

	return nil
}

type GetUserFilter struct {
	Accounts []string
	Status   enum.UserStatus
	Role     enum.UserRole
}

func GetUsers(ctx context.Context, db *sql.DB, filter *GetUserFilter) (res []*model.User, total int64, err error) {
	var sql strings.Builder
	var wheres []string
	var params []any
	sql.WriteString(" SELECT ")
	sql.WriteString("  `id`, ")
	sql.WriteString("  `account`, ")
	sql.WriteString("  `status`, ")
	sql.WriteString("  `role`, ")
	sql.WriteString("  `password`, ")
	sql.WriteString("  `auth_meta`, ")
	sql.WriteString("  `created_at`, ")
	sql.WriteString("  `updated_at` ")
	sql.WriteString("  `updated_at` ")
	sql.WriteString(" FROM `account_db`.`user` ")
	sql.WriteString(" WHERE ")

	if len(filter.Accounts) > 0 {
		var tmp strings.Builder
		tmp.WriteString("  `account` IN (? ")
		for i := range filter.Accounts {
			if i > 0 {
				tmp.WriteString(",? ")
			}
			params = append(params, filter.Accounts[i])
		}
		tmp.WriteString(" ) ")
		wheres = append(wheres, tmp.String())
	}

	if filter.Status != 0 {
		wheres = append(wheres, " `status` = ? ")
		params = append(params, filter.Status)
	}

	if filter.Role != 0 {
		wheres = append(wheres, " `role` = ? ")
		params = append(params, filter.Role)
	}

	if len(wheres) > 0 {
		sql.WriteString(strings.Join(wheres, " AND "))
	}

	sql.WriteString(" ORDER BY `id` ")
	row, err := db.QueryContext(ctx, sql.String(), params...)
	if err != nil {
		return nil, 0, err
	}
	defer row.Close()

	res = make([]*model.User, 0)
	for row.Next() {
		user := new(model.User)
		if err := row.Scan(
			&user.ID,
			&user.Account,
			&user.Status,
			&user.Role,
			&user.Password,
			&user.AuthMeta,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		res = append(res, user)
	}

	return res, total, nil
}
