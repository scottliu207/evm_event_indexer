package model

import (
	"time"
)

type (
	BlockSync struct {
		Address        string
		LastSyncNumber uint64
		LastSyncHash   string
		// LastSyncTimestamp      time.Time
		// LastFinalizedNumber    uint64
		// LastFinalizedTimestamp time.Time
		UpdatedAt time.Time
	}
)
