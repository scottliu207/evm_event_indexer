package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_JWT(t *testing.T) {
	session := NewJWT("secret")
	token, err := session.GenerateToken(1, 2*time.Second)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	t.Logf("token: %s", token)

	// test valid token
	userID, err := session.VerifyToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userID)

	// test invalid token
	userID, err = session.VerifyToken("invalid_token")
	assert.Error(t, err)

	// test valid token with bearer prefix
	userID, err = session.VerifyToken("bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userID)

	// test valid token with Bearer prefix (canonical form)
	userID, err = session.VerifyToken("Bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userID)

	time.Sleep(3 * time.Second)
	_, err = session.VerifyToken(token)
	assert.Error(t, err)

}
