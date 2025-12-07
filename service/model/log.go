package model

import (
	"database/sql/driver"
	"encoding/json"
	"evm_event_indexer/internal/enum"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type (
	Log struct {
		ID             int64
		ChainType      enum.ChainType
		Address        string
		BlockHash      string
		BlockNumber    uint64
		Topics         *Topics
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

	Topics []common.Hash
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
	return json.Marshal(t)
}

func (t *Topics) Array() []common.Hash {
	if t == nil {
		return []common.Hash{}
	}

	return *t
}

// Scan : 實作 sql.Scanner 介面
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

// Value : 實作 driver.Valuer 界面
func (t *DecodedEvent) Value() (driver.Value, error) {
	if t == nil {
		return json.Marshal(&DecodedEvent{})
	}
	return json.Marshal(t)
}
