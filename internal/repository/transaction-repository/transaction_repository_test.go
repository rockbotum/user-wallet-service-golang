package transaction_repository

import (
	"testing"

	"user-wallet-service/internal/model/request-model"
)

func strPtr(s string) *string { return &s }

func TestBuildTransactionFilter(t *testing.T) {
	tests := []struct {
		name         string
		accountID    string
		filter       request_model.TransactionFilter
		wantWhere    string
		wantArgCount int
	}{
		{
			name:         "no filters - only account condition",
			accountID:    "acc-1",
			filter:       request_model.TransactionFilter{},
			wantWhere:    " WHERE account_id = $1",
			wantArgCount: 1,
		},
		{
			name:      "type name filter uses subquery",
			accountID: "acc-1",
			filter: request_model.TransactionFilter{
				Type: strPtr("deposit"),
			},
			wantWhere:    " WHERE account_id = $1 AND type_id = (SELECT id FROM transaction_types WHERE name = $2)",
			wantArgCount: 2,
		},
		{
			name:      "status filter",
			accountID: "acc-1",
			filter: request_model.TransactionFilter{
				Status: strPtr("active"),
			},
			wantWhere:    " WHERE account_id = $1 AND status = $2",
			wantArgCount: 2,
		},
		{
			name:      "date range filters",
			accountID: "acc-1",
			filter: request_model.TransactionFilter{
				DateFrom: strPtr("2026-01-01T00:00:00Z"),
				DateTo:   strPtr("2026-12-31T23:59:59Z"),
			},
			wantWhere:    " WHERE account_id = $1 AND created_at >= $2 AND created_at <= $3",
			wantArgCount: 3,
		},
		{
			name:      "amount range filters",
			accountID: "acc-1",
			filter: request_model.TransactionFilter{
				AmountMin: strPtr("10.00"),
				AmountMax: strPtr("100.50"),
			},
			wantWhere:    " WHERE account_id = $1 AND amount >= $2 AND amount <= $3",
			wantArgCount: 3,
		},
		{
			name:      "all filters combined keep placeholder order",
			accountID: "acc-1",
			filter: request_model.TransactionFilter{
				Type:      strPtr("withdraw"),
				Status:    strPtr("active"),
				DateFrom:  strPtr("2026-01-01T00:00:00Z"),
				DateTo:    strPtr("2026-12-31T23:59:59Z"),
				AmountMin: strPtr("1.00"),
				AmountMax: strPtr("9.99"),
			},
			wantWhere: " WHERE account_id = $1" +
				" AND type_id = (SELECT id FROM transaction_types WHERE name = $2)" +
				" AND status = $3" +
				" AND created_at >= $4" +
				" AND created_at <= $5" +
				" AND amount >= $6" +
				" AND amount <= $7",
			wantArgCount: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			where, args := buildTransactionFilter(tt.accountID, tt.filter)

			if where != tt.wantWhere {
				t.Fatalf("where = %q, want %q", where, tt.wantWhere)
			}
			if len(args) != tt.wantArgCount {
				t.Fatalf("len(args) = %d, want %d", len(args), tt.wantArgCount)
			}
			if len(args) > 0 && args[0] != tt.accountID {
				t.Fatalf("args[0] = %v, want %v", args[0], tt.accountID)
			}
		})
	}
}
