// Package jwt предоставляет менеджер для генерации и валидации JWT токенов.
package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager управляет жизненным циклом JWT токенов.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New создаёт JWT-менеджер с указанными параметрами.
func New(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// AccessClaims содержит данные access-токена.
type AccessClaims struct {
	UserID string `json:"user_id"`
	RoleID int    `json:"role_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken создаёт signed access-токен с указанным TTL.
func (m *Manager) GenerateAccessToken(userID string, roleID int) (string, error) {
	claims := AccessClaims{
		UserID: userID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("jwt: sign access token: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken создаёт случайный hex-токен.
func (m *Manager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("jwt: generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ValidateAccessToken проверяет подпись и срок действия токена, возвращает claims.
func (m *Manager) ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid token: %w", err)
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt: invalid token claims")
	}

	return claims, nil
}

// RefreshTTL возвращает TTL refresh-токена.
func (m *Manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}
