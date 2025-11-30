package hashing_test

import (
	"crypto/rand"
	"encoding/hex"
	"evm_event_indexer/utils/hashing"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Argon2(t *testing.T) {

	a := hashing.NewArgon2(&hashing.Argon2Opt{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 2,
		KeyLen:  32,
	})

	// generate random password
	b := make([]byte, 16)
	rand.Read(b)
	pwd := hex.EncodeToString(b)

	// hashing the password
	encrypted, saltB64, err := a.Hashing(pwd)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEmpty(t, saltB64)

	// verify correct password
	assert.True(t, a.Verify(pwd, saltB64, encrypted))
	// verify incorrect password
	assert.False(t, a.Verify("wrong_password", saltB64, encrypted))

}
