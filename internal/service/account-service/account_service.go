package account_service

import (
	"context"
	"fmt"

	error_model "user-wallet-service/internal/model/error-model"
	response_model "user-wallet-service/internal/model/response-model"
	transaction_model "user-wallet-service/internal/model/transaction-model"
)

// AccountService содержит бизнес-логику работы со счетами.
type AccountService struct {
	accounts     AccountRepository
	transactions TransactionRepository
}

// NewAccountService создаёт AccountService.
func NewAccountService(
	accounts AccountRepository,
	transactions TransactionRepository,
) *AccountService {
	return &AccountService{
		accounts:     accounts,
		transactions: transactions,
	}
}

// GetBalance возвращает баланс аккаунта пользователя.
func (s *AccountService) GetBalance(ctx context.Context, userID string) (*response_model.AccountResponse, error) {
	account, err := s.accounts.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &response_model.AccountResponse{
		ID:      account.ID,
		Balance: account.Balance,
	}, nil
}

// Deposit увеличивает баланс аккаунта.
func (s *AccountService) Deposit(ctx context.Context, userID, amount string) error {
	account, err := s.accounts.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if _, err := s.accounts.UpdateBalance(ctx, account.ID.String(), amount); err != nil {
		return err
	}

	tx := &transaction_model.Transaction{
		AccountID: account.ID,
		TypeID:    transaction_model.TypeIDDeposit,
		Amount:    amount,
		Status:    "active",
	}
	return s.transactions.Create(ctx, tx)
}

// Withdraw уменьшает баланс аккаунта (с проверкой неотрицательного баланса).
func (s *AccountService) Withdraw(ctx context.Context, userID, amount string) error {
	account, err := s.accounts.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// передаём отрицательную дельту для списания
	if _, err := s.accounts.UpdateBalance(ctx, account.ID.String(), "-"+amount); err != nil {
		return err
	}

	tx := &transaction_model.Transaction{
		AccountID: account.ID,
		TypeID:    transaction_model.TypeIDWithdraw,
		Amount:    amount,
		Status:    "active",
	}
	return s.transactions.Create(ctx, tx)
}

// Transfer выполняет атомарный перевод между счетами.
func (s *AccountService) Transfer(ctx context.Context, fromUserID, toUserID, amount string) error {
	if fromUserID == toUserID {
		return error_model.New(error_model.KindInvalid, "cannot transfer to yourself")
	}

	fromAccount, err := s.accounts.FindByUserID(ctx, fromUserID)
	if err != nil {
		return fmt.Errorf("account service: find sender: %w", err)
	}

	toAccount, err := s.accounts.FindByUserID(ctx, toUserID)
	if err != nil {
		return fmt.Errorf("account service: find receiver: %w", err)
	}

	return s.accounts.Transfer(ctx, fromAccount.ID.String(), toAccount.ID.String(), amount)
}
