package auth_controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/mail"

	"user-wallet-service/internal/controller/respond"
	request_model "user-wallet-service/internal/model/request-model"
	auth_service "user-wallet-service/internal/service/auth-service"
)

// AuthController обрабатывает HTTP-запросы аутентификации.
type AuthController struct {
	svc *auth_service.AuthService
	log *slog.Logger
}

// NewAuthController создаёт AuthController.
func NewAuthController(svc *auth_service.AuthService, log *slog.Logger) *AuthController {
	return &AuthController{svc: svc, log: log}
}

// Register обрабатывает POST /auth/register.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req request_model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if err := validateRegister(req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := c.svc.Register(r.Context(), req.Email, req.Password); err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Login обрабатывает POST /auth/login.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req request_model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if err := validateLogin(req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", err)
		return
	}

	resp, err := c.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	respond.WriteJSON(w, http.StatusOK, resp)
}

// Refresh обрабатывает POST /auth/refresh.
func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	var req request_model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if req.RefreshToken == "" {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", map[string]string{"refresh_token": "required"})
		return
	}

	resp, err := c.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	respond.WriteJSON(w, http.StatusOK, resp)
}

// Logout обрабатывает POST /auth/logout.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	var req request_model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if req.RefreshToken == "" {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", map[string]string{"refresh_token": "required"})
		return
	}

	if err := c.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// validateRegister проверяет обязательные поля и минимальную длину пароля.
func validateRegister(req request_model.RegisterRequest) map[string]string {
	errs := make(map[string]string)
	if req.Email == "" {
		errs["email"] = "required"
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		errs["email"] = "invalid email format"
	}
	if req.Password == "" {
		errs["password"] = "required"
	} else if len(req.Password) < 8 {
		errs["password"] = "at least 8 characters"
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// validateLogin проверяет обязательные поля.
func validateLogin(req request_model.LoginRequest) map[string]string {
	errs := make(map[string]string)
	if req.Email == "" {
		errs["email"] = "required"
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		errs["email"] = "invalid email format"
	}
	if req.Password == "" {
		errs["password"] = "required"
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
