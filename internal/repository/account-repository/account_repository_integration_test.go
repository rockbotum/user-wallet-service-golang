//go:build integration

package account_repository

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"user-wallet-service/internal/infrastructure/config"
	"user-wallet-service/internal/infrastructure/database"
	"user-wallet-service/internal/infrastructure/migrate"
	"user-wallet-service/internal/model/error-model"
	"user-wallet-service/internal/model/transaction-model"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration setup: load config: %v\n", err)
		os.Exit(1)
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../../migrations"
	}
	if err := migrate.Run(databaseURL, migrationsPath); err != nil {
		fmt.Fprintf(os.Stderr, "integration setup: migrate: %v\n", err)
		os.Exit(1)
	}

	testDB, err = database.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration setup: connect: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if code != 0 {
		_, _ = testDB.ExecContext(context.Background(),
			`TRUNCATE transactions, sessions, accounts, profiles, users CASCADE`)
	}
	testDB.Close()
	os.Exit(code)
}

// seedAccount создаёт пользователя и аккаунт с указанным балансом, возвращает id аккаунта.
func seedAccount(t *testing.T, email string, balance string) string {
	t.Helper()
	ctx := context.Background()

	var userID string
	err := testDB.QueryRowxContext(ctx,
		`INSERT INTO users (email, password_hash, role_id) VALUES ($1, 'x', 1) RETURNING id`,
		email).Scan(&userID)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	var accountID string
	if balance == "" {
		balance = "0"
	}
	err = testDB.GetContext(ctx, &accountID,
		`INSERT INTO accounts (user_id, balance) VALUES ($1, $2::numeric) RETURNING id`,
		userID, balance)
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	return accountID
}

func cleanupAccounts(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		_, _ = testDB.ExecContext(context.Background(),
			`DELETE FROM accounts WHERE id = $1`, id)
	}
}

func TestTransfer_MovesBalanceExactly(t *testing.T) {
	repo := NewAccountRepository(testDB)
	from := seedAccount(t, fmt.Sprintf("tr-basic-%d@test.io", time.Now().UnixNano()), "100.05")
	to := seedAccount(t, fmt.Sprintf("tr-basic-%d-to@test.io", time.Now().UnixNano()), "0.10")
	defer cleanupAccounts(t, from, to)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.Transfer(ctx, from, to, "50.25"); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	a, err := repo.FindByID(ctx, from)
	if err != nil {
		t.Fatalf("find sender: %v", err)
	}
	if a.Balance != "49.80" {
		t.Fatalf("sender balance = %q, want 49.80 (exact numeric)", a.Balance)
	}

	b, err := repo.FindByID(ctx, to)
	if err != nil {
		t.Fatalf("find receiver: %v", err)
	}
	if b.Balance != "50.35" {
		t.Fatalf("receiver balance = %q, want 50.35 (exact numeric)", b.Balance)
	}

	var n int
	if err := testDB.GetContext(ctx, &n,
		`SELECT COUNT(*) FROM transactions WHERE account_id IN ($1, $2) AND type_id IN ($3, $4)`,
		from, to, transaction_model.TypeIDTransferOut, transaction_model.TypeIDTransferIn); err != nil {
		t.Fatalf("count transfer rows: %v", err)
	}
	if n != 2 {
		t.Fatalf("transfer transaction rows = %d, want 2", n)
	}
}

func TestTransfer_InsufficientFundsRejected(t *testing.T) {
	repo := NewAccountRepository(testDB)
	from := seedAccount(t, fmt.Sprintf("tr-poor-%d@test.io", time.Now().UnixNano()), "5.00")
	to := seedAccount(t, fmt.Sprintf("tr-poor-%d-to@test.io", time.Now().UnixNano()), "0")
	defer cleanupAccounts(t, from, to)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := repo.Transfer(ctx, from, to, "10.00")
	if !error_model.Is(err, error_model.KindInvalid) {
		t.Fatalf("Transfer() error = %v, want KindInvalid insufficient funds", err)
	}

	a, _ := repo.FindByID(ctx, from)
	if a.Balance != "5.00" {
		t.Fatalf("sender balance changed on failed transfer: %q", a.Balance)
	}
}

func TestTransfer_ConcurrentOppositeTransfers_NoDeadlock(t *testing.T) {
	repo := NewAccountRepository(testDB)
	a := seedAccount(t, fmt.Sprintf("tr-race-%d-a@test.io", time.Now().UnixNano()), "200.00")
	b := seedAccount(t, fmt.Sprintf("tr-race-%d-b@test.io", time.Now().UnixNano()), "200.00")
	defer cleanupAccounts(t, a, b)

	var wg sync.WaitGroup
	errs := make([]error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		errs[0] = repo.Transfer(ctx, a, b, "30.00") // A -> B
	}()
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		errs[1] = repo.Transfer(ctx, b, a, "20.00") // B -> A (встречный)
	}()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("concurrent opposite transfers deadlocked")
	}

	for i, err := range errs {
		if err != nil {
			t.Fatalf("opposite transfer %d failed: %v", i, err)
		}
	}

	ctx := context.Background()
	accA, _ := repo.FindByID(ctx, a)
	accB, _ := repo.FindByID(ctx, b)
	if accA.Balance != "190.00" {
		t.Fatalf("account A balance = %q, want 190.00 (-30 +20)", accA.Balance)
	}
	if accB.Balance != "210.00" {
		t.Fatalf("account B balance = %q, want 210.00 (+30 -20)", accB.Balance)
	}
}

func TestUpdateBalance_RejectsNegativeResult(t *testing.T) {
	repo := NewAccountRepository(testDB)
	id := seedAccount(t, fmt.Sprintf("upd-%d@test.io", time.Now().UnixNano()), "3.00")
	defer cleanupAccounts(t, id)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := repo.UpdateBalance(ctx, id, "-10.00"); !error_model.Is(err, error_model.KindInvalid) {
		t.Fatalf("UpdateBalance() error = %v, want KindInvalid", err)
	}

	got, err := repo.UpdateBalance(ctx, id, "-3.00")
	if err != nil {
		t.Fatalf("UpdateBalance(-3.00) error = %v", err)
	}
	if got != "0.00" {
		t.Fatalf("balance = %q, want 0.00", got)
	}
}
