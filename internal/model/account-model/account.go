package account_model

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID     uuid.UUID `db:"id"`
	UserID uuid.UUID `db:"user_id"`
	// Balance — точное строковое представление NUMERIC(21,2), без потерь при сканировании.
	Balance   string     `db:"balance"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
