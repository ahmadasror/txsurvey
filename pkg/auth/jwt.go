// Package auth mints and validates the app's own session token (HS256), issued
// after a successful Google sign-in and carried in an httpOnly cookie. This
// mirrors the txhcs JWTManager convention but is a SESSION minter (single token,
// no refresh) rather than a Bearer consumer — txsurvey is its own identity
// provider for creators.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
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

// JWTManager signs and verifies session tokens with a shared HS256 secret. It
// also keeps an in-memory revocation set (by token id / jti) so that logout
// truly invalidates a session instead of merely dropping the cookie. The set is
// process-local and resets on restart — acceptable at this scale, and revoked
// ids are forgotten once the token's own expiry passes (bounded size).
type JWTManager struct {
	secret  []byte
	ttl     time.Duration
	mu      sync.Mutex
	revoked map[string]time.Time // jti -> token expiry
}

// NewJWTManager builds the session signer. ttl is the session lifetime.
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), ttl: ttl, revoked: make(map[string]time.Time)}
}

// TTL is the configured session lifetime (used to set the cookie MaxAge).
func (m *JWTManager) TTL() time.Duration { return m.ttl }

// GenerateSessionToken issues a signed session token for a creator. Each token
// carries a unique id (jti) so it can be revoked individually on logout.
func (m *JWTManager) GenerateSessionToken(userID, email, name string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newJTI(),
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateSessionToken parses and verifies a session token, rejecting any
// non-HMAC signing method (alg-confusion guard) and any revoked token.
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
	if claims.ID != "" && m.isRevoked(claims.ID) {
		return nil, fmt.Errorf("token revoked")
	}
	return claims, nil
}

// RevokeToken invalidates a still-valid session token (called on logout). A
// malformed/expired token is a no-op. The id is remembered until the token's
// own expiry, after which it can never be replayed anyway.
func (m *JWTManager) RevokeToken(tokenStr string) {
	claims, err := m.ValidateSessionToken(tokenStr)
	if err != nil || claims.ID == "" {
		return
	}
	exp := time.Now().Add(m.ttl)
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}
	m.mu.Lock()
	m.revoked[claims.ID] = exp
	m.mu.Unlock()
}

func (m *JWTManager) isRevoked(jti string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for k, exp := range m.revoked {
		if now.After(exp) {
			delete(m.revoked, k) // forget expired revocations to bound the map
		}
	}
	_, ok := m.revoked[jti]
	return ok
}

func newJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "" // empty jti => simply not individually revocable
	}
	return hex.EncodeToString(b)
}
