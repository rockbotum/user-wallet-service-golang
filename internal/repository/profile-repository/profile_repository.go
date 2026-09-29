package profile_repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"user-wallet-service/internal/model/error-model"
	"user-wallet-service/internal/model/profile-model"
)

type ProfileRepository struct {
	db *sqlx.DB
}

func NewProfileRepository(db *sqlx.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID string) (*profile_model.Profile, error) {
	var profile profile_model.Profile
	query := `SELECT user_id, first_name, last_name, age, created_at, updated_at
		FROM profiles WHERE user_id = $1`

	if err := r.db.QueryRowxContext(ctx, query, userID).StructScan(&profile); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, error_model.Wrap(error_model.KindNotFound, "profile not found", err)
		}
		return nil, error_model.Wrap(error_model.KindInternal, "failed to find profile", err)
	}
	return &profile, nil
}

func (r *ProfileRepository) Upsert(ctx context.Context, profile *profile_model.Profile) error {
	query := `
		INSERT INTO profiles (user_id, first_name, last_name, age)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			age = EXCLUDED.age`

	_, err := r.db.ExecContext(ctx, query,
		profile.UserID, profile.FirstName, profile.LastName, profile.Age)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to upsert profile", err)
	}
	return nil
}
