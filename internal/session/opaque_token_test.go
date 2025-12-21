package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OpaqueToken(t *testing.T) {
	token, err := NewOpaqueToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	t.Logf("token: %s", token)
}
