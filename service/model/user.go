package model

import (
	"database/sql/driver"
	"encoding/json"
	"evm_event_indexer/internal/enum"
	"fmt"
	"time"
)

type (
	User struct {
		ID        int64           // user id
		Account   string          // user account
		Status    enum.UserStatus // user status (enabled, disabled)
		Password  string          // hashed password
		AuthMeta  *AuthMeta       // JSON string of auth metadata
		CreatedAt time.Time       // creation timestamp
		UpdatedAt time.Time       // last update timestamp
	}

	AuthMeta struct {
		Salt    string `json:"salt"`
		Memory  uint32 `json:"memory"`
		Time    uint32 `json:"time"`
		Threads uint8  `json:"threads"`
		KeyLen  uint32 `json:"key_len"`
	}
)

// Scan : implement driver.Scanner interface
func (t *AuthMeta) Scan(val any) error {
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
func (t *AuthMeta) Value() (driver.Value, error) {
	if t == nil {
		return json.Marshal(&AuthMeta{})
	}
	return json.Marshal(t)
}
