package transaction_service

import (
	"testing"

	transaction_model "user-wallet-service/internal/model/transaction-model"
)

func TestTypeName(t *testing.T) {
	tests := []struct {
		id   int
		want string
	}{
		{transaction_model.TypeIDDeposit, "deposit"},
		{transaction_model.TypeIDWithdraw, "withdraw"},
		{transaction_model.TypeIDTransferIn, "transfer_in"},
		{transaction_model.TypeIDTransferOut, "transfer_out"},
		{999, "unknown"},
		{0, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := typeName(tt.id); got != tt.want {
				t.Fatalf("typeName(%d) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
