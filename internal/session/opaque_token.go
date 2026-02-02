package session

import (
	"crypto/rand"
	"encoding/hex"
	"evm_event_indexer/utils/hashing"
)

func NewOpaqueToken() (plain string, hashed string, err error) {
	b := make([]byte, 32) // 256-bit entropy
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	plain = hex.EncodeToString(b)
	hashed = hashing.Sha256([]byte(plain))

	return plain, hashed, nil
}
