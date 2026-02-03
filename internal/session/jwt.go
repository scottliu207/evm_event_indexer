package session

import (
	"errors"
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

type JwtClaim struct {
	Sid string `json:"sid"`
	jwt.RegisteredClaims
}

func NewJWT(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
}

// GenerateToken generates a JWT token for a given user ID and expiration time
func (j *JWT) GenerateToken(userID int64, sessionID string, expiration time.Duration) (string, *jwt.Token, error) {
	if len(j.secret) == 0 {
		return "", nil, fmt.Errorf("jwt secret is empty")
	}
	if expiration <= 0 {
		return "", nil, fmt.Errorf("jwt expiration must be > 0")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &JwtClaim{
		Sid: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
			Issuer:    jwtIssuer,
			Audience:  jwt.ClaimStrings{jwtAudience},
		},
	})

	tokenStr, err := token.SignedString(j.secret)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, token, nil
}

// VerifyToken verifies a JWT token and returns the claims
// If the token is invalid, it returns nil claims and nil error
// error returns only when server errors occur
func (j *JWT) VerifyToken(tokenString string) (*JwtClaim, error) {
	if len(j.secret) == 0 {
		return nil, fmt.Errorf("jwt secret is empty")
	}

	raw := strings.TrimSpace(tokenString)
	if raw == "" {
		return nil, nil
	}

	if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		raw = strings.TrimSpace(raw[len("bearer "):])
	}

	if raw == "" {
		return nil, nil
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	tokenObj, err := parser.ParseWithClaims(raw, new(JwtClaim), func(token *jwt.Token) (any, error) {
		return j.secret, nil
	})
	if err != nil {
		// Non-server errors should be treated as invalid credentials.
		if errors.Is(err, jwt.ErrTokenMalformed) ||
			errors.Is(err, jwt.ErrTokenUnverifiable) ||
			errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenExpired) ||
			errors.Is(err, jwt.ErrTokenNotValidYet) ||
			errors.Is(err, jwt.ErrTokenInvalidAudience) ||
			errors.Is(err, jwt.ErrTokenInvalidIssuer) ||
			errors.Is(err, jwt.ErrTokenInvalidId) ||
			errors.Is(err, jwt.ErrTokenInvalidClaims) ||
			errors.Is(err, jwt.ErrTokenUsedBeforeIssued) {
			return nil, nil
		}
		return nil, err
	}

	if tokenObj == nil {
		return nil, nil
	}

	claims, ok := tokenObj.Claims.(*JwtClaim)
	if !ok {
		return nil, nil
	}

	if !tokenObj.Valid {
		return nil, nil
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, nil
	}

	if claims.Issuer != jwtIssuer {
		return nil, nil
	}

	if !slices.Contains(claims.Audience, jwtAudience) {
		return nil, nil
	}

	if claims.Subject == "" {
		return nil, nil
	}

	return claims, nil
}
