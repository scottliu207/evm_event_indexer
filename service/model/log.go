package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const TableNameEventLog = "event_db.event_log"

type (
	Log struct {
		ID             int64
		ChainID        int64
		Address        string
		BlockHash      string
		BlockNumber    uint64
		Topic0         string
		Topic1         string
		Topic2         string
		Topic3         string
		TxIndex        int32
		LogIndex       int32
		TxHash         string
		Data           []byte
		DecodedEvent   *DecodedEvent
		BlockTimestamp time.Time
		CreatedAt      time.Time
	}

	DecodedEvent struct {
		EventName string            `json:"event_name"`
		EventData map[string]string `json:"event_data"`
	}
)

// Scan : implement sql.Scanner interface
func (t *DecodedEvent) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

// Value : implement driver.Valuer interface
func (t *DecodedEvent) Value() (driver.Value, error) {
	if t == nil {
		return json.Marshal(&DecodedEvent{})
	}
	return json.Marshal(t)
}
