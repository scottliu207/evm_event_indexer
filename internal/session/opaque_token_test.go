package session_test

import (
	"evm_event_indexer/internal/session"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OpaqueToken(t *testing.T) {
	plain, hashed, err := session.NewOpaqueToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, plain)
	assert.NotEmpty(t, hashed)
	t.Logf("plain_token: %s, hashed_token: %s", plain, hashed)
}
