package transaction_model

import (
	"time"

	"github.com/google/uuid"
)

// Идентификаторы типов транзакций; значения синхронизированы с сидом 002_seed_data.up.sql.
const (
	TypeIDDeposit     = 1
	TypeIDWithdraw    = 2
	TypeIDTransferIn  = 3
	TypeIDTransferOut = 4
)

type Transaction struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	TypeID    int       `db:"type_id"`
	// Amount — точное строковое представление NUMERIC(21,2), без потерь при сканировании.
	Amount    string    `db:"amount"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

type TransactionType struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}
