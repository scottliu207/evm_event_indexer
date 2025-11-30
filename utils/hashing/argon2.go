package hashing

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/argon2"
)

type (
	Argon2 struct {
		time    uint32
		memory  uint32
		threads uint8
		keyLen  uint32
	}

	Argon2Opt struct {
		Time    uint32 // Number of iterations
		Memory  uint32 // Memory usage in KB
		Threads uint8  // Number of parallel threads
		KeyLen  uint32 // Length of generated key in bytes
	}
)

/*
default option:
  - time: 1
  - memory: 64 * 1024
  - threads: 2
  - keyLen: 32
*/
func NewArgon2(opt *Argon2Opt) *Argon2 {
	time := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(2)
	keyLen := uint32(32)
	if opt != nil {
		if opt.Time > 0 {
			time = opt.Time
		}
		if opt.Memory > 0 {
			memory = opt.Memory
		}
		if opt.Threads > 0 {
			threads = opt.Threads
		}
		if opt.KeyLen > 0 {
			keyLen = opt.KeyLen
		}
	}

	return &Argon2{
		time:    time,
		memory:  memory,
		threads: threads,
		keyLen:  keyLen,
	}
}

func (a *Argon2) Hashing(pwd string) (string, string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	hash := argon2.IDKey([]byte(pwd), salt, a.time, a.memory, a.threads, a.keyLen)

	return base64.RawStdEncoding.EncodeToString(hash), base64.RawStdEncoding.EncodeToString(salt), nil
}

func (a *Argon2) Verify(pwd string, saltB64 string, hashB64 string) bool {
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false
	}
	oldHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false
	}

	newHash := argon2.IDKey([]byte(pwd), salt, a.time, a.memory, a.threads, a.keyLen)
	if len(newHash) != len(oldHash) {
		return false
	}

	// This ensures constant-time comparison to prevent timing attacks
	var diff byte
	for i := range newHash {
		diff |= newHash[i] ^ oldHash[i]
	}

	return diff == 0
}
