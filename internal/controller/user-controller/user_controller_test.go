package user_controller

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

func TestGetMe_NoContext(t *testing.T) {
	ctrl := NewUserController(nil, noopLog())
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()

	ctrl.GetMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUpdateProfile_InvalidJSON(t *testing.T) {
	ctrl := NewUserController(nil, noopLog())
	req := httptest.NewRequest(http.MethodPatch, "/me/profile", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()

	ctrl.UpdateProfile(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUpdateProfile_NoContext(t *testing.T) {
	ctrl := NewUserController(nil, noopLog())
	body := `{"first_name":"John"}`
	req := httptest.NewRequest(http.MethodPatch, "/me/profile", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.UpdateProfile(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
