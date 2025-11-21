package model

import (
	"time"
)

type (
	BlockSync struct {
		Address                string
		LastSyncNumber         uint64
		LastSyncTimestamp      time.Time
		LastFinalizedNumber    uint64
		LastFinalizedTimestamp time.Time
		UpdatedAt              time.Time
	}
)
