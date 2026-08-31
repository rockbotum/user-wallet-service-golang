package account_service

import (
	"context"

	account_model "user-wallet-service/internal/model/account-model"
	transaction_model "user-wallet-service/internal/model/transaction-model"
)

type AccountRepository interface {
	FindByUserID(ctx context.Context, userID string) (*account_model.Account, error)
	UpdateBalance(ctx context.Context, id string, delta string) (string, error)
	Transfer(ctx context.Context, fromID, toID string, amount string) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *transaction_model.Transaction) error
}
