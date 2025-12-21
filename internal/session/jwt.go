package session

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	jwtIssuer   = "evm_event_indexer"
	jwtAudience = "evm_event_indexer"
)

type JWT struct {
	secret []byte
}

func NewJWT(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
}

// GenerateToken generates a JWT token for a given user ID and expiration time
func (j *JWT) GenerateToken(userID int64, expiration time.Duration) (string, error) {
	if len(j.secret) == 0 {
		return "", fmt.Errorf("jwt secret is empty")
	}
	if expiration <= 0 {
		return "", fmt.Errorf("jwt expiration must be > 0")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.NewString(),
		Issuer:    jwtIssuer,
		Audience:  jwt.ClaimStrings{jwtAudience},
	})
	return token.SignedString(j.secret)
}

// VerifyToken verifies a JWT token and returns the user ID if the token is valid
// If the token is invalid, it returns userID = 0 and nil error
func (j *JWT) VerifyToken(tokenString string) (int64, error) {
	if len(j.secret) == 0 {
		return 0, fmt.Errorf("jwt secret is empty")
	}

	raw := strings.TrimSpace(tokenString)
	if raw == "" {
		return 0, fmt.Errorf("token is empty")
	}

	if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		raw = strings.TrimSpace(raw[len("bearer "):])
	}

	if raw == "" {
		return 0, nil
	}

	claims := new(jwt.RegisteredClaims)
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	tokenObj, err := parser.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return j.secret, nil
	})
	if err != nil {
		return 0, err
	}

	if !tokenObj.Valid {
		return 0, nil
	}

	if claims.Issuer != jwtIssuer {
		return 0, nil
	}

	if len(claims.Audience) == 0 {
		return 0, nil
	}

	if !slices.Contains(claims.Audience, jwtAudience) {
		return 0, nil
	}

	if claims.Subject == "" {
		return 0, nil
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid token subject: %w", err)
	}

	return userID, nil
}
