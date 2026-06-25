// Package auth mints and validates the app's own session token (HS256), issued
// after a successful Google sign-in and carried in an httpOnly cookie. This
// mirrors the txhcs JWTManager convention but is a SESSION minter (single token,
// no refresh) rather than a Bearer consumer — txsurvey is its own identity
// provider for creators.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the payload of a creator session token.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager signs and verifies session tokens with a shared HS256 secret.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager builds the session signer. ttl is the session lifetime.
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), ttl: ttl}
}

// TTL is the configured session lifetime (used to set the cookie MaxAge).
func (m *JWTManager) TTL() time.Duration { return m.ttl }

// GenerateSessionToken issues a signed session token for a creator.
func (m *JWTManager) GenerateSessionToken(userID, email, name string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateSessionToken parses and verifies a session token, rejecting any
// non-HMAC signing method (alg-confusion guard).
func (m *JWTManager) ValidateSessionToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
