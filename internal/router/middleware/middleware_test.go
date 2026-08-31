package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"user-wallet-service/internal/infrastructure/jwt"
	context_model "user-wallet-service/internal/model/context-model"
)

const testSecret = "test-secret-key-for-jwt-manager-32ch"

func testJWT() *jwt.Manager {
	return jwt.New(testSecret, 15*time.Minute, 7*24*time.Hour)
}

// ── WithAuth ────────────────────────────────────────────────────

func TestWithAuth_ValidToken(t *testing.T) {
	m := testJWT()
	token, err := m.GenerateAccessToken("user-uuid-123", 1)
	if err != nil {
		t.Fatal(err)
	}

	var gotUserID string
	var gotRoleID int
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = r.Context().Value(context_model.ContextKeyUserID).(string)
		gotRoleID, _ = r.Context().Value(context_model.ContextKeyRoleID).(int)
		w.WriteHeader(http.StatusOK)
	})

	handler := WithAuth(m, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != "user-uuid-123" {
		t.Fatalf("UserID = %q, want %q", gotUserID, "user-uuid-123")
	}
	if gotRoleID != 1 {
		t.Fatalf("RoleID = %d, want 1", gotRoleID)
	}
}

func TestWithAuth_MissingHeader(t *testing.T) {
	m := testJWT()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := WithAuth(m, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithAuth_InvalidScheme(t *testing.T) {
	m := testJWT()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := WithAuth(m, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithAuth_InvalidToken(t *testing.T) {
	m := testJWT()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := WithAuth(m, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithAuth_ExpiredToken(t *testing.T) {
	m := jwt.New(testSecret, -1*time.Second, 7*24*time.Hour)
	token, _ := m.GenerateAccessToken("user-uuid-123", 1)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := WithAuth(m, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithAuth_TokenWithoutBearer(t *testing.T) {
	m := testJWT()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := WithAuth(m, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Token abc")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// ── WithRoleCheck ───────────────────────────────────────────────

func TestWithRoleCheck_Allowed(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := WithRoleCheck("admin")(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	ctx := context.WithValue(req.Context(), context_model.ContextKeyRoleID, 2) // admin = 2
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWithRoleCheck_Denied(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := WithRoleCheck("admin")(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	ctx := context.WithValue(req.Context(), context_model.ContextKeyRoleID, 1) // user = 1
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestWithRoleCheck_MissingRoleInContext(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called")
	})

	handler := WithRoleCheck("admin")(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithRoleCheck_MultipleAllowed(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := WithRoleCheck("user", "admin")(inner)

	for _, roleID := range []int{1, 2} {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := context.WithValue(req.Context(), context_model.ContextKeyRoleID, roleID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("roleID=%d: status = %d, want %d", roleID, rec.Code, http.StatusOK)
		}
	}
}

// ── WithLogger ──────────────────────────────────────────────────

func TestWithLogger_CallsInner(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	handler := WithLogger(log, inner)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("inner handler was not called")
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestWithLogger_CapturesStatus(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := WithLogger(log, inner)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// ── roleIDToName ────────────────────────────────────────────────

func TestRoleIDToName(t *testing.T) {
	tests := []struct {
		id   int
		want string
	}{
		{1, "user"},
		{2, "admin"},
		{99, ""},
		{0, ""},
	}

	for _, tt := range tests {
		if got := roleIDToName(tt.id); got != tt.want {
			t.Fatalf("roleIDToName(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

// ── RateLimiter ─────────────────────────────────────────────────

func TestRateLimit_AllowsWithinBurst(t *testing.T) {
	rl := NewRateLimiter(100, 5, time.Minute)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.WithRateLimit(inner)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
}

func TestRateLimit_ExceedsBurst(t *testing.T) {
	rl := NewRateLimiter(1, 2, time.Minute)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.WithRateLimit(inner)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if i < 2 && rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
		if i == 2 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusTooManyRequests)
		}
	}
}

func TestRateLimit_PerClient(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Minute)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.WithRateLimit(inner)

	reqA1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqA1.RemoteAddr = "10.0.0.3:1234"
	recA1 := httptest.NewRecorder()
	handler.ServeHTTP(recA1, reqA1)
	if recA1.Code != http.StatusOK {
		t.Fatalf("client A request 1: status = %d, want %d", recA1.Code, http.StatusOK)
	}

	reqB1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqB1.RemoteAddr = "10.0.0.4:1234"
	recB1 := httptest.NewRecorder()
	handler.ServeHTTP(recB1, reqB1)
	if recB1.Code != http.StatusOK {
		t.Fatalf("client B request 1: status = %d, want %d", recB1.Code, http.StatusOK)
	}

	reqA2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqA2.RemoteAddr = "10.0.0.3:1234"
	recA2 := httptest.NewRecorder()
	handler.ServeHTTP(recA2, reqA2)
	if recA2.Code != http.StatusTooManyRequests {
		t.Fatalf("client A request 2: status = %d, want %d", recA2.Code, http.StatusTooManyRequests)
	}
}
