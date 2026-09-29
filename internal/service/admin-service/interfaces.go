package admin_service

import (
	"context"

	role_model "user-wallet-service/internal/model/role-model"
	user_model "user-wallet-service/internal/model/user-model"
)

type UserRepository interface {
	List(ctx context.Context, offset, limit int) ([]user_model.User, error)
	Count(ctx context.Context) (int, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	UpdateRole(ctx context.Context, id string, roleID int) error
}

type RoleRepository interface {
	List(ctx context.Context) ([]role_model.Role, error)
}
