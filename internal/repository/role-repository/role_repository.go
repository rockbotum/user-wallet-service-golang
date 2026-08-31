package role_repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"user-wallet-service/internal/model/error-model"
	"user-wallet-service/internal/model/role-model"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByID(ctx context.Context, id int) (*role_model.Role, error) {
	var role role_model.Role
	query := `SELECT id, name, description, created_at FROM roles WHERE id = $1`

	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, error_model.Wrap(error_model.KindNotFound, "role not found", err)
		}
		return nil, error_model.Wrap(error_model.KindInternal, "failed to find role", err)
	}
	return &role, nil
}

func (r *RoleRepository) List(ctx context.Context) ([]role_model.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles ORDER BY id`

	var roles []role_model.Role
	if err := r.db.SelectContext(ctx, &roles, query); err != nil {
		return nil, error_model.Wrap(error_model.KindInternal, "failed to list roles", err)
	}
	return roles, nil
}
