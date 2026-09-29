package response_model

import (
	"github.com/google/uuid"

	profile_model "user-wallet-service/internal/model/profile-model"
)

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID      uuid.UUID              `json:"id"`
	Email   string                 `json:"email"`
	Role    string                 `json:"role"`
	Profile *profile_model.Profile `json:"profile,omitempty"`
}

type AccountResponse struct {
	ID uuid.UUID `json:"id"`
	// Balance — строка с точным десятичным значением (NUMERIC(21,2)).
	Balance string `json:"balance"`
}

type TransactionResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Amount    string    `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
	Page         int                   `json:"page"`
	Limit        int                   `json:"limit"`
}
