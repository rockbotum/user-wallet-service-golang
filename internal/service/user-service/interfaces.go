package user_service

import (
	"context"

	profile_model "user-wallet-service/internal/model/profile-model"
	role_model "user-wallet-service/internal/model/role-model"
	user_model "user-wallet-service/internal/model/user-model"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*user_model.User, error)
}

type RoleRepository interface {
	FindByID(ctx context.Context, id int) (*role_model.Role, error)
}

type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID string) (*profile_model.Profile, error)
	Upsert(ctx context.Context, profile *profile_model.Profile) error
}
