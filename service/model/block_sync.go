package model

import (
	"time"
)

type (
	BlockSync struct {
		ChainID        int64
		Address        string
		LastSyncNumber uint64
		LastSyncHash   string
		UpdatedAt      time.Time
	}
)
