package jwt

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-key-for-jwt-manager-32ch"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	m := New(testSecret, 15*time.Minute, 7*24*time.Hour)

	token, err := m.GenerateAccessToken("user-123", 1)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}

	claims, err := m.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.RoleID != 1 {
		t.Fatalf("RoleID = %d, want %d", claims.RoleID, 1)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	m := New(testSecret, 15*time.Minute, 7*24*time.Hour)

	_, err := m.ValidateAccessToken("not-a-valid-token")
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	m1 := New(testSecret, 15*time.Minute, 7*24*time.Hour)
	m2 := New("another-secret-key-for-jwt-manager-32", 15*time.Minute, 7*24*time.Hour)

	token, _ := m1.GenerateAccessToken("user-123", 1)

	_, err := m2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for wrong secret")
	}
}

func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	m := New(testSecret, -1*time.Second, 7*24*time.Hour) // TTL в прошлом

	token, err := m.GenerateAccessToken("user-123", 1)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = m.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for expired token")
	}
}

func TestGenerateRefreshToken_Randomness(t *testing.T) {
	m := New(testSecret, 15*time.Minute, 7*24*time.Hour)

	t1, err := m.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	t2, err := m.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if t1 == t2 {
		t.Fatal("two refresh tokens should not be equal")
	}
}

func TestGenerateRefreshToken_Length(t *testing.T) {
	m := New(testSecret, 15*time.Minute, 7*24*time.Hour)

	token, err := m.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	// 32 bytes = 64 hex chars
	if len(token) != 64 {
		t.Fatalf("len(refresh token) = %d, want 64", len(token))
	}
}

func TestGenerateAccessToken_TokenFormat(t *testing.T) {
	m := New(testSecret, 15*time.Minute, 7*24*time.Hour)

	token, _ := m.GenerateAccessToken("user-123", 1)

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token parts = %d, want 3 (header.payload.signature)", len(parts))
	}
}

func TestRefreshTTL(t *testing.T) {
	m := New(testSecret, 15*time.Minute, 7*24*time.Hour)

	if got := m.RefreshTTL(); got != 7*24*time.Hour {
		t.Fatalf("RefreshTTL() = %v, want %v", got, 7*24*time.Hour)
	}
}
