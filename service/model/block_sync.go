package model

import (
	"time"
)

type (
	BlockSync struct {
		ChainID        int64     // chain id
		Address        string    // contract address
		LastSyncNumber uint64    // block number
		LastSyncHash   string    // block hash
		UpdatedAt      time.Time // updated at
	}
)
