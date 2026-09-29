package profile_model

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID    uuid.UUID  `db:"user_id"`
	FirstName *string    `db:"first_name"`
	LastName  *string    `db:"last_name"`
	Age       *int       `db:"age"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
