package transaction_service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	account_model "user-wallet-service/internal/model/account-model"
	error_model "user-wallet-service/internal/model/error-model"
	request_model "user-wallet-service/internal/model/request-model"
	transaction_model "user-wallet-service/internal/model/transaction-model"

	transaction_service "user-wallet-service/internal/service/transaction-service"
	"user-wallet-service/internal/service/transaction-service/mocks"
)

func TestListByAccount(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	accountID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		mockFind    func(ctx context.Context, userID string) (*account_model.Account, error)
		mockList    func(ctx context.Context, accountID string, filter request_model.TransactionFilter) ([]transaction_model.Transaction, int, error)
		wantTotal   int
		wantLen     int
		wantPage    int
		wantLimit   int
		wantErr     bool
		wantErrKind error_model.Kind
	}{
		{
			name: "success",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return &account_model.Account{ID: accountID, Balance: "100.00"}, nil
			},
			mockList: func(_ context.Context, aid string, _ request_model.TransactionFilter) ([]transaction_model.Transaction, int, error) {
				if aid != accountID.String() {
					t.Fatalf("accountID = %s, want %s", aid, accountID)
				}
				return []transaction_model.Transaction{
					{
						ID:        uuid.New(),
						AccountID: accountID,
						TypeID:    transaction_model.TypeIDDeposit,
						Amount:    "100.00",
						Status:    "active",
						CreatedAt: now,
					},
					{
						ID:        uuid.New(),
						AccountID: accountID,
						TypeID:    transaction_model.TypeIDWithdraw,
						Amount:    "25.00",
						Status:    "active",
						CreatedAt: now,
					},
				}, 2, nil
			},
			wantTotal: 2,
			wantLen:   2,
			wantPage:  1,
			wantLimit: 20,
			wantErr:   false,
		},
		{
			name: "account not found",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return nil, error_model.New(error_model.KindNotFound, "account not found")
			},
			mockList: func(_ context.Context, _ string, _ request_model.TransactionFilter) ([]transaction_model.Transaction, int, error) {
				t.Fatal("ListByAccountID should not be called")
				return nil, 0, nil
			},
			wantErr:     true,
			wantErrKind: error_model.KindNotFound,
		},
		{
			name: "empty list",
			mockFind: func(_ context.Context, _ string) (*account_model.Account, error) {
				return &account_model.Account{ID: accountID, Balance: "0.00"}, nil
			},
			mockList: func(_ context.Context, _ string, _ request_model.TransactionFilter) ([]transaction_model.Transaction, int, error) {
				return []transaction_model.Transaction{}, 0, nil
			},
			wantTotal: 0,
			wantLen:   0,
			wantPage:  1,
			wantLimit: 20,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mocks.AccountRepositoryMock{
				FindByUserIDFunc: tt.mockFind,
			}
			txRepo := &mocks.TransactionRepositoryMock{
				ListByAccountIDFunc: tt.mockList,
			}
			svc := transaction_service.NewTransactionService(accountRepo, txRepo)

			resp, err := svc.ListByAccount(ctx, userID, request_model.TransactionFilter{
				Page:  0,
				Limit: 0,
			})
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
			if resp.Total != tt.wantTotal {
				t.Fatalf("total = %d, want %d", resp.Total, tt.wantTotal)
			}
			if len(resp.Transactions) != tt.wantLen {
				t.Fatalf("transactions len = %d, want %d", len(resp.Transactions), tt.wantLen)
			}
			if resp.Page != tt.wantPage {
				t.Fatalf("page = %d, want %d", resp.Page, tt.wantPage)
			}
			if resp.Limit != tt.wantLimit {
				t.Fatalf("limit = %d, want %d", resp.Limit, tt.wantLimit)
			}
			// verify type names for non-empty responses
			if tt.wantLen > 0 {
				if resp.Transactions[0].Type != "deposit" {
					t.Fatalf("first transaction type = %q, want %q", resp.Transactions[0].Type, "deposit")
				}
				if resp.Transactions[1].Type != "withdraw" {
					t.Fatalf("second transaction type = %q, want %q", resp.Transactions[1].Type, "withdraw")
				}
			}
		})
	}
}

func TestListByAccountRepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	accountID := uuid.New()

	accountRepo := &mocks.AccountRepositoryMock{
		FindByUserIDFunc: func(_ context.Context, _ string) (*account_model.Account, error) {
			return &account_model.Account{ID: accountID}, nil
		},
	}
	txRepo := &mocks.TransactionRepositoryMock{
		ListByAccountIDFunc: func(_ context.Context, _ string, _ request_model.TransactionFilter) ([]transaction_model.Transaction, int, error) {
			return nil, 0, fmt.Errorf("db connection lost")
		},
	}
	svc := transaction_service.NewTransactionService(accountRepo, txRepo)

	resp, err := svc.ListByAccount(ctx, userID, request_model.TransactionFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if resp != nil {
		t.Fatal("expected nil response on error")
	}
}
