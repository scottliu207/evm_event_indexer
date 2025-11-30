package model

import (
	"time"
)

type (
	BlockSync struct {
		Address        string
		LastSyncNumber uint64
		LastSyncHash   string
		UpdatedAt      time.Time
	}
)
