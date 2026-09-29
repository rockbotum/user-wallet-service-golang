package admin_service

import (
	"context"

	role_model "user-wallet-service/internal/model/role-model"
)

// AdminService содержит бизнес-логику администраторского управления.
type AdminService struct {
	users UserRepository
	roles RoleRepository
}

// NewAdminService создаёт AdminService.
func NewAdminService(users UserRepository, roles RoleRepository) *AdminService {
	return &AdminService{users: users, roles: roles}
}

// UserListItem — пользователь в списке администратора (без пароля).
type UserListItem struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ListUsers возвращает страницу пользователей с именами ролей.
func (s *AdminService) ListUsers(ctx context.Context, offset, limit int) ([]UserListItem, int, error) {
	users, err := s.users.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.users.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rolesMap, err := s.rolesMap(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]UserListItem, len(users))
	for i, u := range users {
		items[i] = UserListItem{
			ID:        u.ID.String(),
			Email:     u.Email,
			Role:      rolesMap[u.RoleID],
			Status:    u.Status,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return items, total, nil
}

// BlockUser блокирует пользователя по UUID.
func (s *AdminService) BlockUser(ctx context.Context, id string) error {
	return s.users.UpdateStatus(ctx, id, "blocked")
}

// UnblockUser разблокирует пользователя по UUID.
func (s *AdminService) UnblockUser(ctx context.Context, id string) error {
	return s.users.UpdateStatus(ctx, id, "active")
}

// UpdateRole меняет роль пользователя.
func (s *AdminService) UpdateRole(ctx context.Context, userID string, roleID int) error {
	return s.users.UpdateRole(ctx, userID, roleID)
}

// ListRoles возвращает все роли.
func (s *AdminService) ListRoles(ctx context.Context) ([]role_model.Role, error) {
	return s.roles.List(ctx)
}

// rolesMap строит маппинг role_id → name.
func (s *AdminService) rolesMap(ctx context.Context) (map[int]string, error) {
	roles, err := s.roles.List(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, len(roles))
	for _, r := range roles {
		m[r.ID] = r.Name
	}
	return m, nil
}
