package auth_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"user-wallet-service/internal/infrastructure/password"
	account_model "user-wallet-service/internal/model/account-model"
	error_model "user-wallet-service/internal/model/error-model"
	session_model "user-wallet-service/internal/model/session-model"
	user_model "user-wallet-service/internal/model/user-model"
	auth_service "user-wallet-service/internal/service/auth-service"
	"user-wallet-service/internal/service/auth-service/mocks"
)

const testPassword = "testpassword123"

var (
	testUserID    = uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	testSessionID = uuid.MustParse("11111111-2222-3333-4444-555555555555")
)

func testUser() *user_model.User {
	return &user_model.User{
		ID:           testUserID,
		Email:        "test@example.com",
		PasswordHash: mustHash(testPassword),
		RoleID:       1,
		Status:       "active",
	}
}

func blockedUser() *user_model.User {
	u := testUser()
	u.Status = "blocked"
	return u
}

func mustHash(p string) string {
	h, err := password.Hash(p)
	if err != nil {
		panic(err)
	}
	return h
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		email    string
		setup    func() (*mocks.UserRepositoryMock, *mocks.AccountRepositoryMock)
		wantErr  bool
		wantKind error_model.Kind
	}{
		{
			name:  "success - new email creates user and account",
			email: "new@example.com",
			setup: func() (*mocks.UserRepositoryMock, *mocks.AccountRepositoryMock) {
				users := &mocks.UserRepositoryMock{
					ExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) {
						return false, nil
					},
					CreateFunc: func(_ context.Context, _ *user_model.User) error {
						return nil
					},
				}
				accounts := &mocks.AccountRepositoryMock{
					CreateFunc: func(_ context.Context, _ *account_model.Account) error {
						return nil
					},
				}
				return users, accounts
			},
			wantErr: false,
		},
		{
			name:  "conflict - email already exists",
			email: "existing@example.com",
			setup: func() (*mocks.UserRepositoryMock, *mocks.AccountRepositoryMock) {
				users := &mocks.UserRepositoryMock{
					ExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) {
						return true, nil
					},
				}
				return users, nil
			},
			wantErr:  true,
			wantKind: error_model.KindConflict,
		},
		{
			name:  "user repo error on ExistsByEmail",
			email: "err@example.com",
			setup: func() (*mocks.UserRepositoryMock, *mocks.AccountRepositoryMock) {
				users := &mocks.UserRepositoryMock{
					ExistsByEmailFunc: func(_ context.Context, _ string) (bool, error) {
						return false, errors.New("db connection lost")
					},
				}
				return users, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, accounts := tt.setup()
			svc := auth_service.NewAuthService(users, accounts, nil, nil)

			err := svc.Register(ctx, tt.email, testPassword)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantKind != "" && !error_model.Is(err, tt.wantKind) {
					t.Errorf("error kind = %v, want %v", err, tt.wantKind)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(users.CreateCalls()) != 1 {
				t.Errorf("expected 1 User.Create call, got %d", len(users.CreateCalls()))
			}
			if accounts != nil && len(accounts.CreateCalls()) != 1 {
				t.Errorf("expected 1 Account.Create call, got %d", len(accounts.CreateCalls()))
			}
		})
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		email     string
		setup     func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock)
		wantErr   bool
		wantKind  error_model.Kind
		wantToken string
	}{
		{
			name:  "success - valid credentials returns tokens",
			email: "test@example.com",
			setup: func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock) {
				users := &mocks.UserRepositoryMock{
					FindByEmailFunc: func(_ context.Context, _ string) (*user_model.User, error) {
						return testUser(), nil
					},
				}
				sessions := &mocks.SessionRepositoryMock{
					CreateFunc: func(_ context.Context, _ *session_model.Session) error {
						return nil
					},
				}
				jwt := &mocks.JWTManagerMock{
					GenerateAccessTokenFunc: func(userID string, roleID int) (string, error) {
						return "access-token-123", nil
					},
					GenerateRefreshTokenFunc: func() (string, error) {
						return "refresh-token-456", nil
					},
					RefreshTTLFunc: func() time.Duration {
						return 24 * time.Hour
					},
				}
				return users, sessions, jwt
			},
			wantErr:   false,
			wantToken: "access-token-123",
		},
		{
			name:  "user not found",
			email: "missing@example.com",
			setup: func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock) {
				users := &mocks.UserRepositoryMock{
					FindByEmailFunc: func(_ context.Context, _ string) (*user_model.User, error) {
						return nil, error_model.New(error_model.KindNotFound, "user not found")
					},
				}
				return users, nil, nil
			},
			wantErr:  true,
			wantKind: error_model.KindUnauthorized,
		},
		{
			name:  "wrong password",
			email: "test@example.com",
			setup: func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock) {
				users := &mocks.UserRepositoryMock{
					FindByEmailFunc: func(_ context.Context, _ string) (*user_model.User, error) {
						// user exists but with different password hash
						u := testUser()
						u.PasswordHash = mustHash("wrongpassword")
						return u, nil
					},
				}
				return users, nil, nil
			},
			wantErr:  true,
			wantKind: error_model.KindUnauthorized,
		},
		{
			name:  "blocked account",
			email: "blocked@example.com",
			setup: func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock) {
				users := &mocks.UserRepositoryMock{
					FindByEmailFunc: func(_ context.Context, _ string) (*user_model.User, error) {
						return blockedUser(), nil
					},
				}
				return users, nil, nil
			},
			wantErr:  true,
			wantKind: error_model.KindForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, sessions, jwt := tt.setup()
			svc := auth_service.NewAuthService(users, nil, sessions, jwt)

			resp, err := svc.Login(ctx, tt.email, testPassword)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantKind != "" && !error_model.Is(err, tt.wantKind) {
					t.Errorf("error kind = %v, want %v", err, tt.wantKind)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.AccessToken != tt.wantToken {
				t.Errorf("AccessToken = %q, want %q", resp.AccessToken, tt.wantToken)
			}
			if resp.RefreshToken != "refresh-token-456" {
				t.Errorf("RefreshToken = %q, want %q", resp.RefreshToken, "refresh-token-456")
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		token    string
		setup    func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock)
		wantErr  bool
		wantKind error_model.Kind
	}{
		{
			name:  "success - rotates session tokens",
			token: "old-refresh-token",
			setup: func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock) {
				users := &mocks.UserRepositoryMock{
					FindByIDFunc: func(_ context.Context, _ string) (*user_model.User, error) {
						return testUser(), nil
					},
				}
				sessions := &mocks.SessionRepositoryMock{
					FindByRefreshTokenHashFunc: func(_ context.Context, _ string) (*session_model.Session, error) {
						return &session_model.Session{
							ID:               testSessionID,
							UserID:           testUserID,
							RefreshTokenHash: []byte("old-refresh-token"),
							ExpiresAt:        time.Now().Add(time.Hour),
						}, nil
					},
					DeleteByRefreshTokenHashFunc: func(_ context.Context, _ string) error {
						return nil
					},
					CreateFunc: func(_ context.Context, _ *session_model.Session) error {
						return nil
					},
				}
				jwt := &mocks.JWTManagerMock{
					GenerateAccessTokenFunc: func(userID string, roleID int) (string, error) {
						return "new-access-token", nil
					},
					GenerateRefreshTokenFunc: func() (string, error) {
						return "new-refresh-token", nil
					},
					RefreshTTLFunc: func() time.Duration {
						return 24 * time.Hour
					},
				}
				return users, sessions, jwt
			},
			wantErr: false,
		},
		{
			name:  "session not found",
			token: "invalid-token",
			setup: func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock) {
				sessions := &mocks.SessionRepositoryMock{
					FindByRefreshTokenHashFunc: func(_ context.Context, _ string) (*session_model.Session, error) {
						return nil, error_model.New(error_model.KindNotFound, "session not found")
					},
				}
				return nil, sessions, nil
			},
			wantErr:  true,
			wantKind: error_model.KindUnauthorized,
		},
		{
			name:  "blocked user",
			token: "valid-token",
			setup: func() (*mocks.UserRepositoryMock, *mocks.SessionRepositoryMock, *mocks.JWTManagerMock) {
				users := &mocks.UserRepositoryMock{
					FindByIDFunc: func(_ context.Context, _ string) (*user_model.User, error) {
						return blockedUser(), nil
					},
				}
				sessions := &mocks.SessionRepositoryMock{
					FindByRefreshTokenHashFunc: func(_ context.Context, _ string) (*session_model.Session, error) {
						return &session_model.Session{
							ID:               testSessionID,
							UserID:           testUserID,
							RefreshTokenHash: []byte("valid-token"),
							ExpiresAt:        time.Now().Add(time.Hour),
						}, nil
					},
				}
				return users, sessions, nil
			},
			wantErr:  true,
			wantKind: error_model.KindForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, sessions, jwt := tt.setup()
			svc := auth_service.NewAuthService(users, nil, sessions, jwt)

			resp, err := svc.Refresh(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantKind != "" && !error_model.Is(err, tt.wantKind) {
					t.Errorf("error kind = %v, want %v", err, tt.wantKind)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.AccessToken != "new-access-token" {
				t.Errorf("AccessToken = %q, want %q", resp.AccessToken, "new-access-token")
			}
			if resp.RefreshToken != "new-refresh-token" {
				t.Errorf("RefreshToken = %q, want %q", resp.RefreshToken, "new-refresh-token")
			}
			if len(sessions.DeleteByRefreshTokenHashCalls()) != 1 {
				t.Errorf("expected 1 Delete call, got %d", len(sessions.DeleteByRefreshTokenHashCalls()))
			}
			if len(sessions.CreateCalls()) != 1 {
				t.Errorf("expected 1 Create call, got %d", len(sessions.CreateCalls()))
			}
		})
	}
}

func TestLogout(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		token   string
		setup   func() *mocks.SessionRepositoryMock
		wantErr bool
	}{
		{
			name:  "success - deletes session",
			token: "refresh-token-abc",
			setup: func() *mocks.SessionRepositoryMock {
				return &mocks.SessionRepositoryMock{
					DeleteByRefreshTokenHashFunc: func(_ context.Context, _ string) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:  "session not found",
			token: "nonexistent-token",
			setup: func() *mocks.SessionRepositoryMock {
				return &mocks.SessionRepositoryMock{
					DeleteByRefreshTokenHashFunc: func(_ context.Context, _ string) error {
						return error_model.New(error_model.KindNotFound, "session not found")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := tt.setup()
			svc := auth_service.NewAuthService(nil, nil, sessions, nil)

			err := svc.Logout(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(sessions.DeleteByRefreshTokenHashCalls()) != 1 {
				t.Errorf("expected 1 Delete call, got %d", len(sessions.DeleteByRefreshTokenHashCalls()))
			}
			if sessions.DeleteByRefreshTokenHashCalls()[0].TokenHash != tt.token {
				t.Errorf("Delete called with %q, want %q",
					sessions.DeleteByRefreshTokenHashCalls()[0].TokenHash, tt.token)
			}
		})
	}
}
