package auth

import (
	"testing"
	"time"
)

func TestJWTManager_RoundTrip(t *testing.T) {
	m := NewJWTManager("test-secret-at-least-32-characters-long!!", time.Hour)
	tok, err := m.GenerateSessionToken("user-1", "a@b.com", "Alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := m.ValidateSessionToken(tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "a@b.com" || claims.Name != "Alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWTManager_RejectsWrongSecret(t *testing.T) {
	signer := NewJWTManager("test-secret-at-least-32-characters-long!!", time.Hour)
	tok, _ := signer.GenerateSessionToken("user-1", "a@b.com", "Alice")

	other := NewJWTManager("DIFFERENT-secret-at-least-32-characters!!", time.Hour)
	if _, err := other.ValidateSessionToken(tok); err == nil {
		t.Fatal("expected validation to fail with a different secret")
	}
}

func TestJWTManager_RejectsExpired(t *testing.T) {
	m := NewJWTManager("test-secret-at-least-32-characters-long!!", -time.Minute)
	tok, _ := m.GenerateSessionToken("user-1", "a@b.com", "Alice")
	if _, err := m.ValidateSessionToken(tok); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
