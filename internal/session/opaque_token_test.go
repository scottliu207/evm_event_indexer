package session_test

import (
	"evm_event_indexer/internal/session"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OpaqueToken(t *testing.T) {
	token, err := session.NewOpaqueToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	t.Logf("token: %s", token)
}
