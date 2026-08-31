package account_controller

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func noopLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		wantErr bool
	}{
		{"valid integer", "100", false},
		{"valid decimal", "10.50", false},
		{"valid small", "0.01", false},
		{"valid comma", "10,50", false},
		{"empty", "", true},
		{"negative", "-5", true},
		{"too many decimals", "10.123", true},
		{"letters", "abc", true},
		{"special chars", "10$50", true},
		{"leading dot", ".50", true},
		{"single zero", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateAmount(tt.amount)
			if tt.wantErr && errs == nil {
				t.Fatalf("validateAmount(%q) = nil, want error", tt.amount)
			}
			if !tt.wantErr && errs != nil {
				t.Fatalf("validateAmount(%q) = %v, want nil", tt.amount, errs)
			}
		})
	}
}

func TestGetBalance_NoContext(t *testing.T) {
	ctrl := NewAccountController(nil, noopLog())
	req := httptest.NewRequest(http.MethodGet, "/account", nil)
	rec := httptest.NewRecorder()

	ctrl.GetBalance(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDeposit_InvalidJSON(t *testing.T) {
	ctrl := NewAccountController(nil, noopLog())
	body := `{bad`
	req := httptest.NewRequest(http.MethodPost, "/account/deposit", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Deposit(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDeposit_MissingAmount(t *testing.T) {
	ctrl := NewAccountController(nil, noopLog())
	body := `{"amount":""}`
	req := httptest.NewRequest(http.MethodPost, "/account/deposit", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Deposit(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTransfer_MissingFields(t *testing.T) {
	ctrl := NewAccountController(nil, noopLog())
	body := `{"to_user_id":"","amount":""}`
	req := httptest.NewRequest(http.MethodPost, "/account/transfer", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Transfer(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithdraw_NoContext(t *testing.T) {
	ctrl := NewAccountController(nil, noopLog())
	body := `{"amount":"10"}`
	req := httptest.NewRequest(http.MethodPost, "/account/withdraw", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Withdraw(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTransfer_NoContext(t *testing.T) {
	ctrl := NewAccountController(nil, noopLog())
	body := `{"to_user_id":"abc","amount":"10"}`
	req := httptest.NewRequest(http.MethodPost, "/account/transfer", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Transfer(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
