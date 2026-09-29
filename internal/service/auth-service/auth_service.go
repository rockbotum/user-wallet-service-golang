package auth_service

import (
	"context"
	"fmt"
	"time"

	"user-wallet-service/internal/infrastructure/password"
	account_model "user-wallet-service/internal/model/account-model"
	error_model "user-wallet-service/internal/model/error-model"
	response_model "user-wallet-service/internal/model/response-model"
	session_model "user-wallet-service/internal/model/session-model"
	user_model "user-wallet-service/internal/model/user-model"
)

// AuthService инкапсулирует бизнес-логику аутентификации.
type AuthService struct {
	users    UserRepository
	accounts AccountRepository
	sessions SessionRepository
	jwt      JWTManager
}

// NewAuthService создаёт AuthService с указанными зависимостями.
func NewAuthService(
	users UserRepository,
	accounts AccountRepository,
	sessions SessionRepository,
	jwtMgr JWTManager,
) *AuthService {
	return &AuthService{
		users:    users,
		accounts: accounts,
		sessions: sessions,
		jwt:      jwtMgr,
	}
}

// Register создаёт нового пользователя с ролью user и пустым аккаунтом.
// Возвращает ошибку KindConflict, если email уже занят.
func (s *AuthService) Register(ctx context.Context, email, passwordStr string) error {
	exists, err := s.users.ExistsByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("auth: check email: %w", err)
	}
	if exists {
		return error_model.New(error_model.KindConflict, "email already registered")
	}

	hash, err := password.Hash(passwordStr)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}

	user := &user_model.User{
		Email:        email,
		PasswordHash: hash,
		RoleID:       1, // user
		Status:       "active",
	}
	if err := s.users.Create(ctx, user); err != nil {
		return fmt.Errorf("auth: create user: %w", err)
	}

	account := &account_model.Account{UserID: user.ID}
	if err := s.accounts.Create(ctx, account); err != nil {
		return fmt.Errorf("auth: create account: %w", err)
	}

	return nil
}

// Login проверяет учётные данные и возвращает пару токенов.
// Возвращает KindUnauthorized при неверных данных или заблокированном аккаунте.
func (s *AuthService) Login(ctx context.Context, email, passwordStr string) (*response_model.LoginResponse, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, error_model.Wrap(error_model.KindUnauthorized, "invalid credentials", err)
	}

	if user.Status != "active" {
		return nil, error_model.New(error_model.KindForbidden, "account is blocked")
	}

	if err := password.Check(passwordStr, user.PasswordHash); err != nil {
		return nil, error_model.Wrap(error_model.KindUnauthorized, "invalid credentials", err)
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("auth: generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth: generate refresh token: %w", err)
	}

	session := &session_model.Session{
		UserID:           user.ID,
		RefreshTokenHash: []byte(refreshToken),
		ExpiresAt:        time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("auth: create session: %w", err)
	}

	return &response_model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Refresh заменяет текущую сессию на новую пару токенов.
// Старая сессия удаляется (ротация refresh token).
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*response_model.LoginResponse, error) {
	session, err := s.sessions.FindByRefreshTokenHash(ctx, refreshToken)
	if err != nil {
		return nil, error_model.Wrap(error_model.KindUnauthorized, "invalid or expired refresh token", err)
	}

	user, err := s.users.FindByID(ctx, session.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("auth: find user for session: %w", err)
	}

	if user.Status != "active" {
		return nil, error_model.New(error_model.KindForbidden, "account is blocked")
	}

	// ротация: удаляем старую сессию
	if err := s.sessions.DeleteByRefreshTokenHash(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("auth: delete old session: %w", err)
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("auth: generate access token: %w", err)
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth: generate refresh token: %w", err)
	}

	newSession := &session_model.Session{
		UserID:           user.ID,
		RefreshTokenHash: []byte(newRefreshToken),
		ExpiresAt:        time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err := s.sessions.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("auth: create session: %w", err)
	}

	return &response_model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout удаляет сессию по refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.sessions.DeleteByRefreshTokenHash(ctx, refreshToken)
}
