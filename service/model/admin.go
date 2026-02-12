package model

import (
	"time"

	"evm_event_indexer/internal/enum"
)

const TableNameAdmin = "account_db.admin"

type Admin struct {
	ID        int64
	Account   string
	Status    enum.UserStatus
	Password  string
	AuthMeta  *AuthMeta
	CreatedAt time.Time
	UpdatedAt time.Time
}
