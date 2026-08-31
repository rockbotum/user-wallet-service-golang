// Package password предоставляет хеширование и проверку паролей через bcrypt.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

// Hash создаёт bcrypt-хэш пароля.
func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("password: hash: %w", err)
	}
	return string(hash), nil
}

// Check сравнивает пароль с bcrypt-хэшем. Возвращает nil при совпадении.
func Check(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("password: check: %w", err)
	}
	return nil
}
