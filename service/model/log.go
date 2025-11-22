package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type (
	Log struct {
		Address        string
		BlockHash      string
		BlockNumber    uint64
		Topics         *Topics
		TxIndex        uint
		LogIndex       uint
		TxHash         string
		Data           []byte
		BlockTimestamp time.Time
		CreatedAt      time.Time
	}

	Topics []string
)

// Scan : 實作 sql.Scanner 介面
func (t *Topics) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

// Value : 實作 driver.Valuer 界面
func (t *Topics) Value() (driver.Value, error) {
	if t == nil {
		return json.Marshal(&Topics{})
	}
	return json.Marshal(&t)
}
