// Package router регистрирует роуты и middleware chain.
package router

import (
	"log/slog"
	"net/http"

	"user-wallet-service/internal/controller/account-controller"
	"user-wallet-service/internal/controller/admin-controller"
	"user-wallet-service/internal/controller/auth-controller"
	health_controller "user-wallet-service/internal/controller/health"
	"user-wallet-service/internal/controller/transaction-controller"
	"user-wallet-service/internal/controller/user-controller"
	"user-wallet-service/internal/infrastructure/jwt"
	"user-wallet-service/internal/router/middleware"
)

// New собирает HTTP-handler приложения.
func New(
	log *slog.Logger,
	jwtMgr *jwt.Manager,
	authCtrl *auth_controller.AuthController,
	userCtrl *user_controller.UserController,
	accountCtrl *account_controller.AccountController,
	txCtrl *transaction_controller.TransactionController,
	adminCtrl *admin_controller.AdminController,
	healthCtrl *health_controller.HealthController,
	rateLimiter *middleware.RateLimiter,
) http.Handler {
	mux := http.NewServeMux()

	// ── Публичные роуты (без auth). ────────────────────────────────
	mux.HandleFunc("GET /health", healthCtrl.Check)
	mux.HandleFunc("POST /auth/register", authCtrl.Register)
	mux.HandleFunc("POST /auth/login", authCtrl.Login)
	mux.HandleFunc("POST /auth/refresh", authCtrl.Refresh)
	mux.HandleFunc("POST /auth/logout", authCtrl.Logout)

	// ── Защищённые роуты (любой authenticated user). ───────────────
	mux.Handle("GET /me", middleware.WithAuth(jwtMgr, http.HandlerFunc(userCtrl.GetMe)))
	mux.Handle("PATCH /me/profile", middleware.WithAuth(jwtMgr, http.HandlerFunc(userCtrl.UpdateProfile)))

	mux.Handle("GET /account", middleware.WithAuth(jwtMgr, http.HandlerFunc(accountCtrl.GetBalance)))
	mux.Handle("POST /account/deposit", middleware.WithAuth(jwtMgr, http.HandlerFunc(accountCtrl.Deposit)))
	mux.Handle("POST /account/withdraw", middleware.WithAuth(jwtMgr, http.HandlerFunc(accountCtrl.Withdraw)))
	mux.Handle("POST /account/transfer", middleware.WithAuth(jwtMgr, http.HandlerFunc(accountCtrl.Transfer)))

	mux.Handle("GET /transactions", middleware.WithAuth(jwtMgr, http.HandlerFunc(txCtrl.List)))

	// ── Admin-роуты (auth + role=admin). ───────────────────────────
	mux.Handle("GET /admin/users", middleware.WithAuth(jwtMgr, middleware.WithRoleCheck("admin")(http.HandlerFunc(adminCtrl.ListUsers))))
	mux.Handle("PATCH /admin/users/{id}/block", middleware.WithAuth(jwtMgr, middleware.WithRoleCheck("admin")(http.HandlerFunc(adminCtrl.BlockUser))))
	mux.Handle("PATCH /admin/users/{id}/unblock", middleware.WithAuth(jwtMgr, middleware.WithRoleCheck("admin")(http.HandlerFunc(adminCtrl.UnblockUser))))
	mux.Handle("PATCH /admin/users/{id}/role", middleware.WithAuth(jwtMgr, middleware.WithRoleCheck("admin")(http.HandlerFunc(adminCtrl.UpdateRole))))
	mux.Handle("GET /admin/roles", middleware.WithAuth(jwtMgr, middleware.WithRoleCheck("admin")(http.HandlerFunc(adminCtrl.ListRoles))))

	// ── Middleware chain (снаружи → внутрь). ───────────────────────
	var handler http.Handler = mux
	handler = middleware.WithRecovery(log, handler)
	handler = middleware.WithLogger(log, handler)
	handler = middleware.WithRequestID(handler)
	if rateLimiter != nil {
		handler = rateLimiter.WithRateLimit(handler)
	}

	return handler
}
