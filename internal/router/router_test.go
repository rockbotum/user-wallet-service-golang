package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"user-wallet-service/internal/controller/auth-controller"
	"user-wallet-service/internal/infrastructure/jwt"
	"user-wallet-service/internal/router/middleware"
	auth_service "user-wallet-service/internal/service/auth-service"
)

func TestAuthRoutesRegistered(t *testing.T) {
	log := noopLogger()
	j := jwt.New("test-secret-must-be-at-least-32-chars!!", 0, 0)

	svc := auth_service.NewAuthService(nil, nil, nil, j)
	auth := auth_controller.NewAuthController(svc, log)

	handler := New(log, j, auth, nil, nil, nil, nil, nil, nil)

	routes := []struct {
		method string
		path   string
		code   int
	}{
		{"POST", "/auth/register", http.StatusBadRequest},
		{"POST", "/auth/login", http.StatusBadRequest},
		{"POST", "/auth/refresh", http.StatusBadRequest},
		{"POST", "/auth/logout", http.StatusBadRequest},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != rt.code {
				t.Errorf("status = %d, want %d", rec.Code, rt.code)
			}
		})
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := middleware.WithRecovery(noopLogger(), http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.WithRequestID(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-Id") == "" {
		t.Error("X-Request-Id header not set")
	}
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
