package transaction_repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"user-wallet-service/internal/model/error-model"
	"user-wallet-service/internal/model/request-model"
	"user-wallet-service/internal/model/transaction-model"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *transaction_model.Transaction) error {
	query := `INSERT INTO transactions (account_id, type_id, amount, status)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	return r.db.QueryRowxContext(ctx, query,
		tx.AccountID, tx.TypeID, tx.Amount, tx.Status,
	).StructScan(tx)
}

func (r *TransactionRepository) ListByAccountID(ctx context.Context, accountID string, filter request_model.TransactionFilter) ([]transaction_model.Transaction, int, error) {
	where, args := buildTransactionFilter(accountID, filter)

	// count
	countQuery := "SELECT COUNT(*) FROM transactions" + where
	var total int
	if err := r.db.QueryRowxContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, error_model.Wrap(error_model.KindInternal, "failed to count transactions", err)
	}

	// data
	sort := "created_at DESC"
	if filter.Sort == "asc" {
		sort = "created_at ASC"
	}
	if filter.Sort == "amount_desc" {
		sort = "amount DESC"
	}
	if filter.Sort == "amount_asc" {
		sort = "amount ASC"
	}

	dataQuery := fmt.Sprintf(
		`SELECT id, account_id, type_id, amount, status, created_at FROM transactions%s ORDER BY %s OFFSET $%d LIMIT $%d`,
		where, sort, len(args)+1, len(args)+2,
	)

	offset := 0
	page := filter.Page
	limit := filter.Limit
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}

	args = append(args, offset, limit)

	var transactions []transaction_model.Transaction
	if err := r.db.SelectContext(ctx, &transactions, dataQuery, args...); err != nil {
		return nil, 0, error_model.Wrap(error_model.KindInternal, "failed to list transactions", err)
	}

	return transactions, total, nil
}

func buildTransactionFilter(accountID string, filter request_model.TransactionFilter) (string, []any) {
	var conditions []string
	args := []any{accountID}
	argIdx := 2 // $1 is account_id

	conditions = append(conditions, "account_id = $1")

	if filter.Type != nil {
		// filter.Type содержит имя типа (deposit/withdraw/...), маппим на id параметризованным подзапросом
		conditions = append(conditions, fmt.Sprintf(
			"type_id = (SELECT id FROM transaction_types WHERE name = $%d)", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}
	if filter.AmountMin != nil {
		conditions = append(conditions, fmt.Sprintf("amount >= $%d", argIdx))
		args = append(args, *filter.AmountMin)
		argIdx++
	}
	if filter.AmountMax != nil {
		conditions = append(conditions, fmt.Sprintf("amount <= $%d", argIdx))
		args = append(args, *filter.AmountMax)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	return where, args
}
