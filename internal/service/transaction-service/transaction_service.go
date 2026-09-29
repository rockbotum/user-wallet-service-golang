package transaction_service

import (
	"context"

	request_model "user-wallet-service/internal/model/request-model"
	response_model "user-wallet-service/internal/model/response-model"
	transaction_model "user-wallet-service/internal/model/transaction-model"
)

// TransactionService содержит бизнес-логику работы с транзакциями.
type TransactionService struct {
	accounts AccountRepository
	txRepo   TransactionRepository
}

// NewTransactionService создаёт TransactionService.
func NewTransactionService(
	accounts AccountRepository,
	txRepo TransactionRepository,
) *TransactionService {
	return &TransactionService{accounts: accounts, txRepo: txRepo}
}

// ListByAccount возвращает список транзакций аккаунта с фильтрацией и пагинацией.
func (s *TransactionService) ListByAccount(ctx context.Context, userID string, filter request_model.TransactionFilter) (*response_model.TransactionListResponse, error) {
	account, err := s.accounts.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	transactions, total, err := s.txRepo.ListByAccountID(ctx, account.ID.String(), filter)
	if err != nil {
		return nil, err
	}

	items := make([]response_model.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		items[i] = response_model.TransactionResponse{
			ID:        tx.ID,
			Type:      typeName(tx.TypeID),
			Amount:    tx.Amount,
			Status:    tx.Status,
			CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	return &response_model.TransactionListResponse{
		Transactions: items,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}

// typeName преобразует type_id в строковое имя.
func typeName(id int) string {
	switch id {
	case transaction_model.TypeIDDeposit:
		return "deposit"
	case transaction_model.TypeIDWithdraw:
		return "withdraw"
	case transaction_model.TypeIDTransferIn:
		return "transfer_in"
	case transaction_model.TypeIDTransferOut:
		return "transfer_out"
	default:
		return "unknown"
	}
}
