package respond

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	error_model "user-wallet-service/internal/model/error-model"
	problem_model "user-wallet-service/internal/model/problem-model"
)

func TestWriteProblemMapping(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name   string
		err    error
		status int
		title  string
	}{
		{name: "invalid", err: error_model.New(error_model.KindInvalid, "bad input"), status: http.StatusBadRequest, title: "Validation failed"},
		{name: "unauthorized", err: error_model.New(error_model.KindUnauthorized, "auth required"), status: http.StatusUnauthorized, title: "Authentication required"},
		{name: "forbidden", err: error_model.New(error_model.KindForbidden, "no rights"), status: http.StatusForbidden, title: "Access denied"},
		{name: "not found", err: error_model.New(error_model.KindNotFound, "missing"), status: http.StatusNotFound, title: "Resource not found"},
		{name: "conflict", err: error_model.New(error_model.KindConflict, "duplicate"), status: http.StatusConflict, title: "Resource conflict"},
		{name: "rate limit", err: error_model.New(error_model.KindRateLimit, "too many"), status: http.StatusTooManyRequests, title: "Rate limit exceeded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			WriteProblem(w, r, log, tt.err)

			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}

			body := w.Body.String()
			if !strings.Contains(body, `"type":"`) {
				t.Fatal("missing type field")
			}
			if !strings.Contains(body, `"title":"`) {
				t.Fatal("missing title field")
			}
			if !strings.Contains(body, `"status":`) {
				t.Fatal("missing status field")
			}
			if !strings.Contains(body, `"detail":"`) {
				t.Fatal("missing detail field")
			}
			if !strings.Contains(body, `"instance":"/test"`) {
				t.Fatal("missing instance field")
			}
		})
	}
}

func TestWriteProblemContentType(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	WriteProblem(w, r, log, error_model.New(error_model.KindInvalid, "bad"))

	ct := w.Header().Get("Content-Type")
	if ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
}

func TestWriteProblemHidesInternalDetails(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	WriteProblem(w, r, log, errors.New("db password=secret"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if strings.Contains(w.Body.String(), "secret") {
		t.Fatalf("internal error details leaked: %s", w.Body.String())
	}
}

func TestWriteProblemInsufficient(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/account/transfer", nil)
	w := httptest.NewRecorder()

	WriteProblemInsufficient(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(w.Body.String(), problem_model.ProbTypeInsufficient) {
		t.Fatal("missing insufficient funds type")
	}
}

func TestWriteProblemWithErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	w := httptest.NewRecorder()

	errs := map[string]string{
		"email":    "invalid email format",
		"password": "too short",
	}
	WriteProblemWithErrors(w, r, http.StatusUnprocessableEntity, "Validation failed", errs)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(w.Body.String(), `"email"`) {
		t.Fatal("missing email field in errors")
	}
	if !strings.Contains(w.Body.String(), `"password"`) {
		t.Fatal("missing password field in errors")
	}
}
