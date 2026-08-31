package account_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"user-wallet-service/internal/model/account-model"
	"user-wallet-service/internal/model/error-model"
	"user-wallet-service/internal/model/transaction-model"
)

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *account_model.Account) error {
	query := `INSERT INTO accounts (user_id) VALUES ($1) RETURNING id, balance, created_at, updated_at`

	if err := r.db.QueryRowxContext(ctx, query, account.UserID).StructScan(account); err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to create account", err)
	}
	return nil
}

func (r *AccountRepository) FindByUserID(ctx context.Context, userID string) (*account_model.Account, error) {
	var account account_model.Account
	query := `SELECT id, user_id, balance, created_at, updated_at FROM accounts WHERE user_id = $1`

	if err := r.db.QueryRowxContext(ctx, query, userID).StructScan(&account); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, error_model.Wrap(error_model.KindNotFound, "account not found", err)
		}
		return nil, error_model.Wrap(error_model.KindInternal, "failed to find account by user id", err)
	}
	return &account, nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id string) (*account_model.Account, error) {
	var account account_model.Account
	query := `SELECT id, user_id, balance, created_at, updated_at FROM accounts WHERE id = $1`

	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&account); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, error_model.Wrap(error_model.KindNotFound, "account not found", err)
		}
		return nil, error_model.Wrap(error_model.KindInternal, "failed to find account by id", err)
	}
	return &account, nil
}

// UpdateBalance изменяет баланс на дельту (положительную или отрицательную).
// Дельта — строка с точным десятичным значением; условие гарантирует неотрицательный баланс.
func (r *AccountRepository) UpdateBalance(ctx context.Context, id string, delta string) (string, error) {
	query := `
		UPDATE accounts
		SET balance = balance + $2::numeric
		WHERE id = $1 AND balance + $2::numeric >= 0
		RETURNING balance`

	var newBalance string
	if err := r.db.QueryRowxContext(ctx, query, id, delta).Scan(&newBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", error_model.Wrap(error_model.KindInvalid, "insufficient funds or account not found", nil)
		}
		return "", error_model.Wrap(error_model.KindInternal, "failed to update balance", err)
	}
	return newBalance, nil
}

// Transfer выполняет перевод между счетами в одной SQL-транзакции:
// обе строки блокируются FOR UPDATE в детерминированном порядке (ORDER BY id),
// что исключает deadlock при одновременных встречных переводах.
func (r *AccountRepository) Transfer(ctx context.Context, fromID, toID string, amount string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer tx.Rollback()

	// lock both accounts in a single statement, deterministic order by id
	var lockedIDs []string
	err = tx.SelectContext(ctx,
		&lockedIDs,
		`SELECT id FROM accounts WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`,
		fromID, toID)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to lock accounts", err)
	}

	found := make(map[string]bool, len(lockedIDs))
	for _, id := range lockedIDs {
		found[id] = true
	}
	if !found[fromID] {
		return error_model.New(error_model.KindNotFound, "sender account not found")
	}
	if !found[toID] {
		return error_model.New(error_model.KindNotFound, "receiver account not found")
	}

	// debit sender
	res, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - $2::numeric WHERE id = $1 AND balance >= $2::numeric`,
		fromID, amount)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to debit sender", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return error_model.New(error_model.KindInvalid, "insufficient funds")
	}

	// credit receiver
	_, err = tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + $2::numeric WHERE id = $1`,
		toID, amount)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to credit receiver", err)
	}

	// insert transactions
	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (account_id, type_id, amount) VALUES ($1, $3, $2::numeric)`,
		fromID, amount, transaction_model.TypeIDTransferOut)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to insert transfer_out transaction", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (account_id, type_id, amount) VALUES ($1, $3, $2::numeric)`,
		toID, amount, transaction_model.TypeIDTransferIn)
	if err != nil {
		return error_model.Wrap(error_model.KindInternal, "failed to insert transfer_in transaction", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: commit tx: %w", err)
	}
	return nil
}

func (r *AccountRepository) List(ctx context.Context, offset, limit int) ([]account_model.Account, error) {
	query := `SELECT id, user_id, balance, created_at, updated_at
		FROM accounts ORDER BY created_at DESC OFFSET $1 LIMIT $2`

	var accounts []account_model.Account
	if err := r.db.SelectContext(ctx, &accounts, query, offset, limit); err != nil {
		return nil, error_model.Wrap(error_model.KindInternal, "failed to list accounts", err)
	}
	return accounts, nil
}

func (r *AccountRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowxContext(ctx, `SELECT COUNT(*) FROM accounts`).Scan(&count); err != nil {
		return 0, error_model.Wrap(error_model.KindInternal, "failed to count accounts", err)
	}
	return count, nil
}
