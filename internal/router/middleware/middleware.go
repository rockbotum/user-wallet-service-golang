package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"user-wallet-service/internal/infrastructure/jwt"
	context_model "user-wallet-service/internal/model/context-model"
)

// WithRequestID генерирует уникальный ID для каждого запроса, если он не передан клиентом.
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), context_model.ContextKeyRequestID, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithLogger логирует.method, path, status и длительность каждого запроса.
func WithLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		log.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start).String(),
			"request_id", r.Context().Value(context_model.ContextKeyRequestID),
		)
	})
}

// statusWriter перехватывает статус-код для логирования.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// WithAuth извлекает и валидирует JWT access token из заголовка Authorization.
// При успехе кладёт user_id и role_id в контекст.
func WithAuth(manager *jwt.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, `{"type":"https://example.com/problems/unauthorized","title":"Authentication required","status":401,"detail":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := manager.ValidateAccessToken(tokenStr)
		if err != nil {
			http.Error(w, `{"type":"https://example.com/problems/unauthorized","title":"Authentication required","status":401,"detail":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), context_model.ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, context_model.ContextKeyRoleID, claims.RoleID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithRoleCheck проверяет, что роль пользователя входит в список разрешённых.
func WithRoleCheck(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleID, ok := r.Context().Value(context_model.ContextKeyRoleID).(int)
			if !ok {
				http.Error(w, `{"type":"https://example.com/problems/unauthorized","title":"Authentication required","status":401,"detail":"role not found in context"}`, http.StatusUnauthorized)
				return
			}

			roleName := roleIDToName(roleID)
			for _, allowed := range allowedRoles {
				if roleName == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"type":"https://example.com/problems/forbidden","title":"Access denied","status":403,"detail":"insufficient permissions"}`, http.StatusForbidden)
		})
	}
}

// roleIDToName преобразует numeric role_id в строковое имя (seed-значения).
func roleIDToName(id int) string {
	switch id {
	case 1:
		return "user"
	case 2:
		return "admin"
	default:
		return ""
	}
}

// RateLimiter ограничивает число запросов на ключ (по умолчанию — IP клиента).
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
	ttl      time.Duration
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter создаёт RateLimiter с указанным лимитом и окном жизни записей.
func NewRateLimiter(rps float64, burst int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    rate.Limit(rps),
		burst:    burst,
		ttl:      ttl,
	}
}

// WithRateLimit ограничивает частоту запросов по IP, отвечая 429 при превышении.
func (rl *RateLimiter) WithRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)

		l := rl.limiterFor(key)

		if !l.Allow() {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"type":"https://example.com/problems/rate-limit-exceeded","title":"Rate limit exceeded","status":429,"detail":"too many requests","instance":"` + r.URL.Path + `"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) limiterFor(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v, ok := rl.visitors[key]; ok {
		v.lastSeen = time.Now()
		return v.limiter
	}

	l := rate.NewLimiter(rl.limit, rl.burst)
	rl.visitors[key] = &visitor{limiter: l, lastSeen: time.Now()}

	if len(rl.visitors) > 1000 {
		rl.cleanup()
	}

	return l
}

// cleanup удаляет записи, не посещавшиеся дольше ttl.
func (rl *RateLimiter) cleanup() {
	for key, v := range rl.visitors {
		if time.Since(v.lastSeen) > rl.ttl {
			delete(rl.visitors, key)
		}
	}
}

// clientIP извлекает IP клиента из заголовка X-Forwarded-For либо RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, _ := strings.Cut(r.RemoteAddr, ":")
	return host
}

// WithRecovery оборачивает handler и перехватывает паники, отвечая 500 вместо падения процесса.
func WithRecovery(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered", "value", rec)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
