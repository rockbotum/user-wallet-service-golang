package user_model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	RoleID       int        `db:"role_id"`
	Status       string     `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

// MustParseUUID парсит строку в uuid.UUID; паника при ошибке (для внутренних вызовов).
func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}
