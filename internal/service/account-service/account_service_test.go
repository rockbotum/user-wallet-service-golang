package account_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	account_model "user-wallet-service/internal/model/account-model"
	error_model "user-wallet-service/internal/model/error-model"
	transaction_model "user-wallet-service/internal/model/transaction-model"

	account_service "user-wallet-service/internal/service/account-service"
	"user-wallet-service/internal/service/account-service/mocks"
)

func TestGetBalance(t *testing.T) {
	ctx := context.Background()
	accountID := uuid.New()
	userUUID := uuid.New()
	userID := userUUID.String()

	tests := []struct {
		name        string
		mockFind    func(ctx context.Context, userID string) (*account_model.Account, error)
		wantBalance string
		wantErr     bool
		wantErrKind error_model.Kind
	}{
		{
			name: "success",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return &account_model.Account{
					ID:      accountID,
					UserID:  userUUID,
					Balance: "100.50",
				}, nil
			},
			wantBalance: "100.50",
			wantErr:     false,
		},
		{
			name: "account not found",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return nil, error_model.New(error_model.KindNotFound, "account not found")
			},
			wantErr:     true,
			wantErrKind: error_model.KindNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mocks.AccountRepositoryMock{
				FindByUserIDFunc: tt.mockFind,
			}
			svc := account_service.NewAccountService(accountRepo, &mocks.TransactionRepositoryMock{})

			resp, err := svc.GetBalance(ctx, userID)
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
			if resp.Balance != tt.wantBalance {
				t.Fatalf("balance = %q, want %q", resp.Balance, tt.wantBalance)
			}
		})
	}
}

func TestDeposit(t *testing.T) {
	ctx := context.Background()
	accountID := uuid.New()
	userID := uuid.New().String()

	tests := []struct {
		name        string
		mockFind    func(ctx context.Context, userID string) (*account_model.Account, error)
		mockUpdate  func(ctx context.Context, id string, delta string) (string, error)
		mockCreate  func(ctx context.Context, tx *transaction_model.Transaction) error
		wantErr     bool
		wantErrKind error_model.Kind
	}{
		{
			name: "success",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return &account_model.Account{ID: accountID, Balance: "100.00"}, nil
			},
			mockUpdate: func(_ context.Context, id string, delta string) (string, error) {
				if id != accountID.String() {
					t.Fatalf("UpdateBalance called with wrong id: %s", id)
				}
				if delta != "50.00" {
					t.Fatalf("UpdateBalance delta = %q, want %q", delta, "50.00")
				}
				return "150.00", nil
			},
			mockCreate: func(_ context.Context, tx *transaction_model.Transaction) error {
				if tx.AccountID != accountID {
					t.Fatal("Create called with wrong account ID")
				}
				if tx.TypeID != transaction_model.TypeIDDeposit {
					t.Fatalf("type_id = %d, want %d", tx.TypeID, transaction_model.TypeIDDeposit)
				}
				if tx.Amount != "50.00" {
					t.Fatalf("amount = %q, want %q", tx.Amount, "50.00")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "account not found",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return nil, error_model.New(error_model.KindNotFound, "account not found")
			},
			mockUpdate: func(_ context.Context, _ string, _ string) (string, error) {
				t.Fatal("UpdateBalance should not be called")
				return "", nil
			},
			mockCreate: func(_ context.Context, _ *transaction_model.Transaction) error {
				t.Fatal("Create should not be called")
				return nil
			},
			wantErr:     true,
			wantErrKind: error_model.KindNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mocks.AccountRepositoryMock{
				FindByUserIDFunc:  tt.mockFind,
				UpdateBalanceFunc: tt.mockUpdate,
			}
			txRepo := &mocks.TransactionRepositoryMock{
				CreateFunc: tt.mockCreate,
			}
			svc := account_service.NewAccountService(accountRepo, txRepo)

			err := svc.Deposit(ctx, userID, "50.00")
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
		})
	}
}

func TestWithdraw(t *testing.T) {
	ctx := context.Background()
	accountID := uuid.New()
	userID := uuid.New().String()

	tests := []struct {
		name        string
		mockFind    func(ctx context.Context, userID string) (*account_model.Account, error)
		mockUpdate  func(ctx context.Context, id string, delta string) (string, error)
		mockCreate  func(ctx context.Context, tx *transaction_model.Transaction) error
		wantErr     bool
		wantErrKind error_model.Kind
	}{
		{
			name: "success",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return &account_model.Account{ID: accountID, Balance: "200.00"}, nil
			},
			mockUpdate: func(_ context.Context, id string, delta string) (string, error) {
				if delta != "-30.00" {
					t.Fatalf("UpdateBalance delta = %q, want %q", delta, "-30.00")
				}
				return "170.00", nil
			},
			mockCreate: func(_ context.Context, tx *transaction_model.Transaction) error {
				if tx.TypeID != transaction_model.TypeIDWithdraw {
					t.Fatalf("type_id = %d, want %d", tx.TypeID, transaction_model.TypeIDWithdraw)
				}
				if tx.Amount != "30.00" {
					t.Fatalf("amount = %q, want %q", tx.Amount, "30.00")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "insufficient funds",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return &account_model.Account{ID: accountID, Balance: "10.00"}, nil
			},
			mockUpdate: func(_ context.Context, _ string, _ string) (string, error) {
				return "", error_model.New(error_model.KindInvalid, "insufficient funds")
			},
			mockCreate: func(_ context.Context, _ *transaction_model.Transaction) error {
				t.Fatal("Create should not be called")
				return nil
			},
			wantErr:     true,
			wantErrKind: error_model.KindInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mocks.AccountRepositoryMock{
				FindByUserIDFunc:  tt.mockFind,
				UpdateBalanceFunc: tt.mockUpdate,
			}
			txRepo := &mocks.TransactionRepositoryMock{
				CreateFunc: tt.mockCreate,
			}
			svc := account_service.NewAccountService(accountRepo, txRepo)

			err := svc.Withdraw(ctx, userID, "30.00")
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
		})
	}
}

func TestTransfer(t *testing.T) {
	ctx := context.Background()
	fromUserID := uuid.New().String()
	toUserID := uuid.New().String()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	tests := []struct {
		name         string
		fromUserID   string
		toUserID     string
		mockFind     func(ctx context.Context, userID string) (*account_model.Account, error)
		mockTransfer func(ctx context.Context, fromID string, toID string, amount string) error
		wantErr      bool
		wantErrKind  error_model.Kind
	}{
		{
			name:       "success",
			fromUserID: fromUserID,
			toUserID:   toUserID,
			mockFind: func(_ context.Context, uid string) (*account_model.Account, error) {
				if uid == fromUserID {
					return &account_model.Account{ID: fromAccountID, Balance: "500.00"}, nil
				}
				return &account_model.Account{ID: toAccountID, Balance: "100.00"}, nil
			},
			mockTransfer: func(_ context.Context, fromID string, toID string, amount string) error {
				if fromID != fromAccountID.String() {
					t.Fatalf("fromID = %s, want %s", fromID, fromAccountID)
				}
				if toID != toAccountID.String() {
					t.Fatalf("toID = %s, want %s", toID, toAccountID)
				}
				if amount != "250.00" {
					t.Fatalf("amount = %q, want %q", amount, "250.00")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:       "self-transfer returns invalid",
			fromUserID: fromUserID,
			toUserID:   fromUserID,
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return &account_model.Account{ID: fromAccountID}, nil
			},
			mockTransfer: func(_ context.Context, _, _ string, _ string) error {
				t.Fatal("Transfer should not be called for self-transfer")
				return nil
			},
			wantErr:     true,
			wantErrKind: error_model.KindInvalid,
		},
		{
			name:       "sender not found",
			fromUserID: fromUserID,
			toUserID:   toUserID,
			mockFind: func(_ context.Context, uid string) (*account_model.Account, error) {
				if uid == fromUserID {
					return nil, error_model.New(error_model.KindNotFound, "sender not found")
				}
				return &account_model.Account{ID: toAccountID}, nil
			},
			mockTransfer: func(_ context.Context, _, _ string, _ string) error {
				t.Fatal("Transfer should not be called")
				return nil
			},
			wantErr:     true,
			wantErrKind: error_model.KindNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mocks.AccountRepositoryMock{
				FindByUserIDFunc: tt.mockFind,
				TransferFunc:     tt.mockTransfer,
			}
			svc := account_service.NewAccountService(accountRepo, &mocks.TransactionRepositoryMock{})

			err := svc.Transfer(ctx, tt.fromUserID, tt.toUserID, "250.00")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrKind != "" {
					var domainErr *error_model.Error
					if errors.As(err, &domainErr) {
						if domainErr.Kind != tt.wantErrKind {
							t.Fatalf("error kind = %q, want %q", domainErr.Kind, tt.wantErrKind)
						}
					} else {
						t.Fatalf("expected domain error with kind %q, got: %v", tt.wantErrKind, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
