package user_service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	error_model "user-wallet-service/internal/model/error-model"
	profile_model "user-wallet-service/internal/model/profile-model"
	request_model "user-wallet-service/internal/model/request-model"
	role_model "user-wallet-service/internal/model/role-model"
	user_model "user-wallet-service/internal/model/user-model"

	user_service "user-wallet-service/internal/service/user-service"
	"user-wallet-service/internal/service/user-service/mocks"
)

func strPtr(s string) *string { return &s }

func TestGetByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	roleID := 1

	tests := []struct {
		name         string
		mockUser     func(ctx context.Context, id string) (*user_model.User, error)
		mockRole     func(ctx context.Context, id int) (*role_model.Role, error)
		mockProfile  func(ctx context.Context, userID string) (*profile_model.Profile, error)
		wantEmail    string
		wantRoleName string
		wantProfile  bool
		wantErr      bool
		wantErrKind  error_model.Kind
	}{
		{
			name: "success with profile",
			mockUser: func(_ context.Context, _ string) (*user_model.User, error) {
				return &user_model.User{
					ID:     uuid.MustParse(userID),
					Email:  "test@example.com",
					RoleID: roleID,
					Status: "active",
				}, nil
			},
			mockRole: func(_ context.Context, _ int) (*role_model.Role, error) {
				return &role_model.Role{ID: roleID, Name: "user"}, nil
			},
			mockProfile: func(_ context.Context, _ string) (*profile_model.Profile, error) {
				firstName := "John"
				lastName := "Doe"
				return &profile_model.Profile{
					FirstName: &firstName,
					LastName:  &lastName,
				}, nil
			},
			wantEmail:    "test@example.com",
			wantRoleName: "user",
			wantProfile:  true,
			wantErr:      false,
		},
		{
			name: "success without profile",
			mockUser: func(_ context.Context, _ string) (*user_model.User, error) {
				return &user_model.User{
					ID:     uuid.MustParse(userID),
					Email:  "noprov@example.com",
					RoleID: roleID,
				}, nil
			},
			mockRole: func(_ context.Context, _ int) (*role_model.Role, error) {
				return &role_model.Role{ID: roleID, Name: "user"}, nil
			},
			mockProfile: func(_ context.Context, _ string) (*profile_model.Profile, error) {
				return nil, error_model.New(error_model.KindNotFound, "profile not found")
			},
			wantEmail:    "noprov@example.com",
			wantRoleName: "user",
			wantProfile:  false,
			wantErr:      false,
		},
		{
			name: "user not found",
			mockUser: func(_ context.Context, _ string) (*user_model.User, error) {
				return nil, error_model.New(error_model.KindNotFound, "user not found")
			},
			mockRole: func(_ context.Context, _ int) (*role_model.Role, error) {
				t.Fatal("Role.FindByID should not be called")
				return nil, nil
			},
			mockProfile: func(_ context.Context, _ string) (*profile_model.Profile, error) {
				t.Fatal("Profile.FindByUserID should not be called")
				return nil, nil
			},
			wantErr:     true,
			wantErrKind: error_model.KindNotFound,
		},
		{
			name: "role not found",
			mockUser: func(_ context.Context, _ string) (*user_model.User, error) {
				return &user_model.User{
					ID:     uuid.MustParse(userID),
					Email:  "test@example.com",
					RoleID: roleID,
				}, nil
			},
			mockRole: func(_ context.Context, _ int) (*role_model.Role, error) {
				return nil, fmt.Errorf("role not found")
			},
			mockProfile: func(_ context.Context, _ string) (*profile_model.Profile, error) {
				t.Fatal("Profile.FindByUserID should not be called")
				return nil, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mocks.UserRepositoryMock{FindByIDFunc: tt.mockUser}
			roleRepo := &mocks.RoleRepositoryMock{FindByIDFunc: tt.mockRole}
			profileRepo := &mocks.ProfileRepositoryMock{FindByUserIDFunc: tt.mockProfile}
			svc := user_service.NewUserService(userRepo, roleRepo, profileRepo)

			resp, err := svc.GetByID(ctx, userID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrKind != "" && !error_model.Is(err, tt.wantErrKind) {
					t.Fatalf("expected error kind %q, got %q", tt.wantErrKind, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Email != tt.wantEmail {
				t.Fatalf("email = %q, want %q", resp.Email, tt.wantEmail)
			}
			if resp.Role != tt.wantRoleName {
				t.Fatalf("role = %q, want %q", resp.Role, tt.wantRoleName)
			}
			if tt.wantProfile && resp.Profile == nil {
				t.Fatal("expected profile to be set, got nil")
			}
			if !tt.wantProfile && resp.Profile != nil {
				t.Fatal("expected profile to be nil, got non-nil")
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	now := time.Now()

	tests := []struct {
		name       string
		req        request_model.ProfileUpdateRequest
		mockUpsert func(ctx context.Context, profile *profile_model.Profile) error
		wantErr    bool
	}{
		{
			name: "success without birth date",
			req: request_model.ProfileUpdateRequest{
				FirstName: strPtr("John"),
				LastName:  strPtr("Doe"),
			},
			mockUpsert: func(_ context.Context, p *profile_model.Profile) error {
				if p.FirstName == nil || *p.FirstName != "John" {
					t.Fatalf("FirstName = %v, want John", p.FirstName)
				}
				if p.LastName == nil || *p.LastName != "Doe" {
					t.Fatalf("LastName = %v, want Doe", p.LastName)
				}
				if p.Age != nil {
					t.Fatalf("Age = %v, want nil", p.Age)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "success with birth date",
			req: request_model.ProfileUpdateRequest{
				FirstName: strPtr("Jane"),
				LastName:  strPtr("Smith"),
				BirthDate: &now,
			},
			mockUpsert: func(_ context.Context, p *profile_model.Profile) error {
				if p.Age == nil {
					t.Fatal("Age should not be nil when BirthDate is provided")
				}
				// compute expected age inline to avoid importing unexported calculateAge
				expectedAge := now.Year() - now.Year()
				if now.YearDay() < now.YearDay() {
					expectedAge--
				}
				// birth date == now, so expected age == 0
				if *p.Age != expectedAge {
					t.Fatalf("Age = %d, want %d", *p.Age, expectedAge)
				}
				if p.FirstName == nil || *p.FirstName != "Jane" {
					t.Fatalf("FirstName = %v, want Jane", p.FirstName)
				}
				if p.LastName == nil || *p.LastName != "Smith" {
					t.Fatalf("LastName = %v, want Smith", p.LastName)
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mocks.UserRepositoryMock{}
			roleRepo := &mocks.RoleRepositoryMock{}
			profileRepo := &mocks.ProfileRepositoryMock{UpsertFunc: tt.mockUpsert}
			svc := user_service.NewUserService(userRepo, roleRepo, profileRepo)

			err := svc.UpdateProfile(ctx, userID, tt.req)
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
