package auth_service

import (
	"context"
	"time"

	account_model "user-wallet-service/internal/model/account-model"
	session_model "user-wallet-service/internal/model/session-model"
	user_model "user-wallet-service/internal/model/user-model"
)

type UserRepository interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *user_model.User) error
	FindByEmail(ctx context.Context, email string) (*user_model.User, error)
	FindByID(ctx context.Context, id string) (*user_model.User, error)
}

type AccountRepository interface {
	Create(ctx context.Context, account *account_model.Account) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *session_model.Session) error
	FindByRefreshTokenHash(ctx context.Context, tokenHash string) (*session_model.Session, error)
	DeleteByRefreshTokenHash(ctx context.Context, tokenHash string) error
}

type JWTManager interface {
	GenerateAccessToken(userID string, roleID int) (string, error)
	GenerateRefreshToken() (string, error)
	RefreshTTL() time.Duration
}
