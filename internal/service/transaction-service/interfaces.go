package transaction_service

import (
	"context"

	account_model "user-wallet-service/internal/model/account-model"
	request_model "user-wallet-service/internal/model/request-model"
	transaction_model "user-wallet-service/internal/model/transaction-model"
)

type AccountRepository interface {
	FindByUserID(ctx context.Context, userID string) (*account_model.Account, error)
}

type TransactionRepository interface {
	ListByAccountID(ctx context.Context, accountID string, filter request_model.TransactionFilter) ([]transaction_model.Transaction, int, error)
}
