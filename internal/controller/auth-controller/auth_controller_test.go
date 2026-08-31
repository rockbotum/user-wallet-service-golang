package auth_controller

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	request_model "user-wallet-service/internal/model/request-model"
)

func noopLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name    string
		req     request_model.RegisterRequest
		wantErr bool
	}{
		{"valid", request_model.RegisterRequest{Email: "a@b.com", Password: "12345678"}, false},
		{"valid complex", request_model.RegisterRequest{Email: "user.name+tag@domain.co", Password: "12345678"}, false},
		{"missing email", request_model.RegisterRequest{Password: "12345678"}, true},
		{"missing password", request_model.RegisterRequest{Email: "a@b.com"}, true},
		{"short password", request_model.RegisterRequest{Email: "a@b.com", Password: "123"}, true},
		{"empty both", request_model.RegisterRequest{}, true},
		{"invalid email no at", request_model.RegisterRequest{Email: "userexample.com", Password: "12345678"}, true},
		{"invalid email no domain", request_model.RegisterRequest{Email: "user@", Password: "12345678"}, true},
		{"invalid email spaces", request_model.RegisterRequest{Email: "user @example.com", Password: "12345678"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateRegister(tt.req)
			if tt.wantErr && errs == nil {
				t.Fatal("expected validation errors, got nil")
			}
			if !tt.wantErr && errs != nil {
				t.Fatalf("expected nil, got errors: %v", errs)
			}
		})
	}
}

func TestValidateRegister_PasswordEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"exactly 8 chars", "12345678", false},
		{"7 chars", "1234567", true},
		{"very long", strings.Repeat("a", 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateRegister(request_model.RegisterRequest{Email: "a@b.com", Password: tt.password})
			if tt.wantErr && errs == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && errs != nil {
				t.Fatalf("expected nil, got %v", errs)
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		req     request_model.LoginRequest
		wantErr bool
	}{
		{"valid", request_model.LoginRequest{Email: "a@b.com", Password: "pass"}, false},
		{"missing email", request_model.LoginRequest{Password: "pass"}, true},
		{"missing password", request_model.LoginRequest{Email: "a@b.com"}, true},
		{"empty both", request_model.LoginRequest{}, true},
		{"invalid email", request_model.LoginRequest{Email: "not-an-email", Password: "pass"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateLogin(tt.req)
			if tt.wantErr && errs == nil {
				t.Fatal("expected validation errors, got nil")
			}
			if !tt.wantErr && errs != nil {
				t.Fatalf("expected nil, got errors: %v", errs)
			}
		})
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	ctrl := NewAuthController(nil, noopLog())
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()

	ctrl.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	ctrl := NewAuthController(nil, noopLog())
	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "email") {
		t.Fatal("response should mention email field")
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	ctrl := NewAuthController(nil, noopLog())
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()

	ctrl.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	ctrl := NewAuthController(nil, noopLog())
	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRefresh_EmptyToken(t *testing.T) {
	ctrl := NewAuthController(nil, noopLog())
	body := `{"refresh_token":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Refresh(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLogout_EmptyToken(t *testing.T) {
	ctrl := NewAuthController(nil, noopLog())
	body := `{"refresh_token":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctrl.Logout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegister_InvalidJSON_ContentType(t *testing.T) {
	ctrl := NewAuthController(nil, noopLog())
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()

	ctrl.Register(rec, req)

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/problem+json") {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
}
