package admin_service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	role_model "user-wallet-service/internal/model/role-model"
	user_model "user-wallet-service/internal/model/user-model"

	admin_service "user-wallet-service/internal/service/admin-service"
	"user-wallet-service/internal/service/admin-service/mocks"
)

func TestListUsers(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		mockList  func(ctx context.Context, offset, limit int) ([]user_model.User, error)
		mockCount func(ctx context.Context) (int, error)
		wantTotal int
		wantLen   int
		wantErr   bool
	}{
		{
			name: "success",
			mockList: func(_ context.Context, _, _ int) ([]user_model.User, error) {
				return []user_model.User{
					{
						ID:        uuid.New(),
						Email:     "user1@example.com",
						RoleID:    1,
						Status:    "active",
						CreatedAt: createdAt,
					},
					{
						ID:        uuid.New(),
						Email:     "user2@example.com",
						RoleID:    2,
						Status:    "blocked",
						CreatedAt: createdAt,
					},
				}, nil
			},
			mockCount: func(_ context.Context) (int, error) {
				return 2, nil
			},
			wantTotal: 2,
			wantLen:   2,
			wantErr:   false,
		},
		{
			name: "user repo error",
			mockList: func(_ context.Context, _, _ int) ([]user_model.User, error) {
				return nil, fmt.Errorf("db error")
			},
			mockCount: func(_ context.Context) (int, error) {
				return 0, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mocks.UserRepositoryMock{
				ListFunc:  tt.mockList,
				CountFunc: tt.mockCount,
			}
			roleRepo := &mocks.RoleRepositoryMock{
				ListFunc: func(_ context.Context) ([]role_model.Role, error) {
					return []role_model.Role{
						{ID: 1, Name: "user"},
						{ID: 2, Name: "admin"},
					}, nil
				},
			}
			svc := admin_service.NewAdminService(userRepo, roleRepo)

			items, total, err := svc.ListUsers(ctx, 0, 10)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if total != tt.wantTotal {
				t.Fatalf("total = %d, want %d", total, tt.wantTotal)
			}
			if len(items) != tt.wantLen {
				t.Fatalf("items len = %d, want %d", len(items), tt.wantLen)
			}
		})
	}
}

func TestBlockUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	tests := []struct {
		name       string
		mockStatus func(ctx context.Context, id string, status string) error
		wantErr    bool
	}{
		{
			name: "success",
			mockStatus: func(_ context.Context, id string, status string) error {
				if id != userID {
					t.Fatalf("id = %s, want %s", id, userID)
				}
				if status != "blocked" {
					t.Fatalf("status = %q, want %q", status, "blocked")
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mocks.UserRepositoryMock{
				UpdateStatusFunc: tt.mockStatus,
			}
			svc := admin_service.NewAdminService(userRepo, &mocks.RoleRepositoryMock{})

			err := svc.BlockUser(ctx, userID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUnblockUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	tests := []struct {
		name       string
		mockStatus func(ctx context.Context, id string, status string) error
		wantErr    bool
	}{
		{
			name: "success",
			mockStatus: func(_ context.Context, id string, status string) error {
				if id != userID {
					t.Fatalf("id = %s, want %s", id, userID)
				}
				if status != "active" {
					t.Fatalf("status = %q, want %q", status, "active")
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mocks.UserRepositoryMock{
				UpdateStatusFunc: tt.mockStatus,
			}
			svc := admin_service.NewAdminService(userRepo, &mocks.RoleRepositoryMock{})

			err := svc.UnblockUser(ctx, userID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateRole(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	roleID := 2

	tests := []struct {
		name       string
		mockUpdate func(ctx context.Context, id string, roleID int) error
		wantErr    bool
	}{
		{
			name: "success",
			mockUpdate: func(_ context.Context, id string, rid int) error {
				if id != userID {
					t.Fatalf("id = %s, want %s", id, userID)
				}
				if rid != roleID {
					t.Fatalf("roleID = %d, want %d", rid, roleID)
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mocks.UserRepositoryMock{
				UpdateRoleFunc: tt.mockUpdate,
			}
			svc := admin_service.NewAdminService(userRepo, &mocks.RoleRepositoryMock{})

			err := svc.UpdateRole(ctx, userID, roleID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestListRoles(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		mockList func(ctx context.Context) ([]role_model.Role, error)
		wantLen  int
		wantErr  bool
	}{
		{
			name: "success",
			mockList: func(_ context.Context) ([]role_model.Role, error) {
				return []role_model.Role{
					{ID: 1, Name: "user", Description: "Regular user"},
					{ID: 2, Name: "admin", Description: "Administrator"},
				}, nil
			},
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mocks.RoleRepositoryMock{
				ListFunc: tt.mockList,
			}
			svc := admin_service.NewAdminService(&mocks.UserRepositoryMock{}, roleRepo)

			roles, err := svc.ListRoles(ctx)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(roles) != tt.wantLen {
				t.Fatalf("roles len = %d, want %d", len(roles), tt.wantLen)
			}
		})
	}
}
