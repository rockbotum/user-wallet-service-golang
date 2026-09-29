// Package main — точка входа приложения.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"user-wallet-service/internal/controller/account-controller"
	"user-wallet-service/internal/controller/admin-controller"
	"user-wallet-service/internal/controller/auth-controller"
	health_controller "user-wallet-service/internal/controller/health"
	"user-wallet-service/internal/controller/transaction-controller"
	"user-wallet-service/internal/controller/user-controller"
	"user-wallet-service/internal/infrastructure/config"
	"user-wallet-service/internal/infrastructure/database"
	"user-wallet-service/internal/infrastructure/httpserver"
	"user-wallet-service/internal/infrastructure/jwt"
	"user-wallet-service/internal/infrastructure/logger"
	account_repository "user-wallet-service/internal/repository/account-repository"
	profile_repository "user-wallet-service/internal/repository/profile-repository"
	role_repository "user-wallet-service/internal/repository/role-repository"
	session_repository "user-wallet-service/internal/repository/session-repository"
	transaction_repository "user-wallet-service/internal/repository/transaction-repository"
	user_repository "user-wallet-service/internal/repository/user-repository"
	"user-wallet-service/internal/router"
	"user-wallet-service/internal/router/middleware"
	account_service "user-wallet-service/internal/service/account-service"
	admin_service "user-wallet-service/internal/service/admin-service"
	auth_service "user-wallet-service/internal/service/auth-service"
	transaction_service "user-wallet-service/internal/service/transaction-service"
	user_service "user-wallet-service/internal/service/user-service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(os.Stdout, slog.LevelInfo)

	db, err := database.New(cfg)
	if err != nil {
		log.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Репозитории.
	userRepo := user_repository.NewUserRepository(db)
	accountRepo := account_repository.NewAccountRepository(db)
	sessionRepo := session_repository.NewSessionRepository(db)
	roleRepo := role_repository.NewRoleRepository(db)
	profileRepo := profile_repository.NewProfileRepository(db)
	txRepo := transaction_repository.NewTransactionRepository(db)

	// JWT-менеджер.
	jwtMgr := jwt.New(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	// Сервисы.
	authSvc := auth_service.NewAuthService(userRepo, accountRepo, sessionRepo, jwtMgr)
	userSvc := user_service.NewUserService(userRepo, roleRepo, profileRepo)
	accountSvc := account_service.NewAccountService(accountRepo, txRepo)
	txSvc := transaction_service.NewTransactionService(accountRepo, txRepo)
	adminSvc := admin_service.NewAdminService(userRepo, roleRepo)

	// Контроллеры.
	authCtrl := auth_controller.NewAuthController(authSvc, log)
	userCtrl := user_controller.NewUserController(userSvc, log)
	accountCtrl := account_controller.NewAccountController(accountSvc, log)
	txCtrl := transaction_controller.NewTransactionController(txSvc, log)
	adminCtrl := admin_controller.NewAdminController(adminSvc, log)
	healthCtrl := health_controller.NewHealthController(db)

	// Роутер с middleware chain.
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst, cfg.RateLimitTTL)
	handler := router.New(log, jwtMgr, authCtrl, userCtrl, accountCtrl, txCtrl, adminCtrl, healthCtrl, rateLimiter)

	srv := httpserver.New(handler, httpserver.Config{
		Port:         cfg.HTTPPort,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		log.Info("http server started", "port", cfg.HTTPPort)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		log.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
}
