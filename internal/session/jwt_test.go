package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_JWT(t *testing.T) {
	session := NewJWT("secret")
	token, err := session.GenerateToken(1, 5*time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	t.Logf("token: %s", token)

	// test valid token
	userID, err := session.VerifyToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userID)

	// test invalid token
	_, err = session.VerifyToken("invalid_token")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), userID)

	// test valid token with bearer prefix
	userID, err = session.VerifyToken("bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userID)
}
