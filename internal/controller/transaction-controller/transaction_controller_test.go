package transaction_controller

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func noopLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestTransactionList_NoContext(t *testing.T) {
	ctrl := NewTransactionController(nil, noopLog())
	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()

	ctrl.List(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestParseTransactionFilter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/transactions?page=2&limit=10&sort=asc&type=deposit&status=active&date_from=2026-01-01&date_to=2026-12-31&amount_min=5&amount_max=100", nil)
	filter := parseTransactionFilter(req)

	if filter.Page != 2 {
		t.Fatalf("Page = %d, want 2", filter.Page)
	}
	if filter.Limit != 10 {
		t.Fatalf("Limit = %d, want 10", filter.Limit)
	}
	if filter.Sort != "asc" {
		t.Fatalf("Sort = %q, want %q", filter.Sort, "asc")
	}
	if filter.Type == nil || *filter.Type != "deposit" {
		t.Fatalf("Type = %v, want deposit", filter.Type)
	}
	if filter.Status == nil || *filter.Status != "active" {
		t.Fatalf("Status = %v, want active", filter.Status)
	}
	if filter.DateFrom == nil || *filter.DateFrom != "2026-01-01" {
		t.Fatalf("DateFrom = %v, want 2026-01-01", filter.DateFrom)
	}
	if filter.DateTo == nil || *filter.DateTo != "2026-12-31" {
		t.Fatalf("DateTo = %v, want 2026-12-31", filter.DateTo)
	}
	if filter.AmountMin == nil || *filter.AmountMin != "5" {
		t.Fatalf("AmountMin = %v, want 5", filter.AmountMin)
	}
	if filter.AmountMax == nil || *filter.AmountMax != "100" {
		t.Fatalf("AmountMax = %v, want 100", filter.AmountMax)
	}
}

func TestParseTransactionFilter_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	filter := parseTransactionFilter(req)

	if filter.Page != 1 {
		t.Fatalf("Page = %d, want 1 (default)", filter.Page)
	}
	if filter.Limit != 20 {
		t.Fatalf("Limit = %d, want 20 (default)", filter.Limit)
	}
	if filter.Type != nil {
		t.Fatalf("Type should be nil")
	}
}

func TestDefaultInt(t *testing.T) {
	tests := []struct {
		s        string
		fallback int
		want     int
	}{
		{"", 10, 10},
		{"5", 10, 5},
		{"0", 10, 10},
		{"-1", 10, 10},
		{"abc", 10, 10},
		{"3", 0, 3},
	}

	for _, tt := range tests {
		if got := defaultInt(tt.s, tt.fallback); got != tt.want {
			t.Fatalf("defaultInt(%q, %d) = %d, want %d", tt.s, tt.fallback, got, tt.want)
		}
	}
}
