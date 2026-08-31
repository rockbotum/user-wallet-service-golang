package user_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"user-wallet-service/internal/model/error-model"
	"user-wallet-service/internal/model/user-model"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *user_model.User) error {
	query := `
		INSERT INTO users (email, password_hash, role_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		user.Email, user.PasswordHash, user.RoleID, user.Status,
	).StructScan(user)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user_model.User, error) {
	var user user_model.User
	query := `SELECT id, email, password_hash, role_id, status, created_at, updated_at FROM users WHERE email = $1`

	if err := r.db.QueryRowxContext(ctx, query, email).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, error_model.Wrap(error_model.KindNotFound, "user not found", err)
		}
		return nil, error_model.Wrap(error_model.KindInternal, "failed to find user by email", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user_model.User, error) {
	var user user_model.User
	query := `SELECT id, email, password_hash, role_id, status, created_at, updated_at FROM users WHERE id = $1`

	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, error_model.Wrap(error_model.KindNotFound, "user not found", err)
		}
		return nil, error_model.Wrap(error_model.KindInternal, "failed to find user by id", err)
	}
	return &user, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE users SET status = $2 WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to update user status", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return error_model.Wrap(error_model.KindNotFound, "user not found", nil)
	}
	return nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id string, roleID int) error {
	query := `UPDATE users SET role_id = $2 WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id, roleID)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to update user role", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return error_model.Wrap(error_model.KindNotFound, "user not found", nil)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]user_model.User, error) {
	query := `SELECT id, email, password_hash, role_id, status, created_at, updated_at
		FROM users ORDER BY created_at DESC OFFSET $1 LIMIT $2`

	var users []user_model.User
	if err := r.db.SelectContext(ctx, &users, query, offset, limit); err != nil {
		return nil, error_model.Wrap(error_model.KindInternal, "failed to list users", err)
	}
	return users, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`

	if err := r.db.QueryRowxContext(ctx, query).Scan(&count); err != nil {
		return 0, error_model.Wrap(error_model.KindInternal, "failed to count users", err)
	}
	return count, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	if err := r.db.QueryRowxContext(ctx, query, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("repository: check email exists: %w", err)
	}
	return exists, nil
}
