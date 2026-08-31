package session_repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"user-wallet-service/internal/model/error-model"
	"user-wallet-service/internal/model/session-model"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *session_model.Session) error {
	hash := sha256.Sum256(session.RefreshTokenHash)
	query := `INSERT INTO sessions (user_id, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3) RETURNING id, created_at`

	return r.db.QueryRowxContext(ctx, query,
		session.UserID, hash[:], session.ExpiresAt,
	).StructScan(session)
}

func (r *SessionRepository) FindByRefreshTokenHash(ctx context.Context, token string) (*session_model.Session, error) {
	hash := sha256.Sum256([]byte(token))
	var session session_model.Session
	query := `SELECT id, user_id, refresh_token_hash, expires_at, created_at
		FROM sessions WHERE refresh_token_hash = $1 AND expires_at > $2`

	if err := r.db.QueryRowxContext(ctx, query, hash[:], time.Now()).StructScan(&session); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, error_model.Wrap(error_model.KindUnauthorized, "session not found or expired", err)
		}
		return nil, error_model.Wrap(error_model.KindInternal, "failed to find session", err)
	}
	return &session, nil
}

func (r *SessionRepository) DeleteByRefreshTokenHash(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	query := `DELETE FROM sessions WHERE refresh_token_hash = $1`

	_, err := r.db.ExecContext(ctx, query, hash[:])
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to delete session", err)
	}
	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	res, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, error_model.Wrap(error_model.KindInternal, "failed to delete expired sessions", err)
	}
	return res.RowsAffected()
}
