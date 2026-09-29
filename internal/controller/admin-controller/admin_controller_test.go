package admin_controller

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

func TestUpdateRole_InvalidJSON(t *testing.T) {
	ctrl := NewAdminController(nil, noopLog())
	req := httptest.NewRequest(http.MethodPatch, "/admin/users/123/role", strings.NewReader("{bad"))
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	ctrl.UpdateRole(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateRole_ZeroRoleID(t *testing.T) {
	ctrl := NewAdminController(nil, noopLog())
	body := `{"role_id":0}`
	req := httptest.NewRequest(http.MethodPatch, "/admin/users/123/role", strings.NewReader(body))
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	ctrl.UpdateRole(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBlockUser_EmptyID(t *testing.T) {
	ctrl := NewAdminController(nil, noopLog())
	req := httptest.NewRequest(http.MethodPatch, "/admin/users//block", nil)
	rec := httptest.NewRecorder()

	ctrl.BlockUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
