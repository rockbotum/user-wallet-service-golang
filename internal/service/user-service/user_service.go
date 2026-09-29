package user_service

import (
	"context"
	"fmt"
	"time"

	error_model "user-wallet-service/internal/model/error-model"
	profile_model "user-wallet-service/internal/model/profile-model"
	request_model "user-wallet-service/internal/model/request-model"
	response_model "user-wallet-service/internal/model/response-model"
	user_model "user-wallet-service/internal/model/user-model"
)

// UserService содержит бизнес-логику работы с пользователями и профилями.
type UserService struct {
	users   UserRepository
	roles   RoleRepository
	profile ProfileRepository
}

// NewUserService создаёт UserService.
func NewUserService(
	users UserRepository,
	roles RoleRepository,
	profile ProfileRepository,
) *UserService {
	return &UserService{users: users, roles: roles, profile: profile}
}

// GetByID возвращает пользователя по UUID с профилем и именем роли.
func (s *UserService) GetByID(ctx context.Context, id string) (*response_model.UserResponse, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	role, err := s.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("user service: find role: %w", err)
	}

	resp := &response_model.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  role.Name,
	}

	profile, err := s.profile.FindByUserID(ctx, id)
	if err != nil && !error_model.Is(err, error_model.KindNotFound) {
		return nil, fmt.Errorf("user service: find profile: %w", err)
	}
	if profile != nil {
		resp.Profile = profile
	}

	return resp, nil
}

// UpdateProfile обновляет профиль текущего пользователя (upsert).
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req request_model.ProfileUpdateRequest) error {
	profile := &profile_model.Profile{
		UserID:    user_model.MustParseUUID(userID),
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if req.BirthDate != nil {
		age := calculateAge(*req.BirthDate)
		profile.Age = &age
	}

	return s.profile.Upsert(ctx, profile)
}

// calculateAge вычисляет возраст по дате рождения.
func calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}
