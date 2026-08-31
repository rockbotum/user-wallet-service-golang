package user_controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"user-wallet-service/internal/controller/respond"
	context_model "user-wallet-service/internal/model/context-model"
	error_model "user-wallet-service/internal/model/error-model"
	request_model "user-wallet-service/internal/model/request-model"
	user_service "user-wallet-service/internal/service/user-service"
)

// UserController обрабатывает HTTP-запросы текущего пользователя.
type UserController struct {
	svc *user_service.UserService
	log *slog.Logger
}

// NewUserController создаёт UserController.
func NewUserController(svc *user_service.UserService, log *slog.Logger) *UserController {
	return &UserController{svc: svc, log: log}
}

// GetMe обрабатывает GET /me.
func (c *UserController) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(context_model.ContextKeyUserID).(string)
	if !ok {
		respond.WriteProblem(w, r, c.log, error_model.New(error_model.KindUnauthorized, "user not authenticated"))
		return
	}

	resp, err := c.svc.GetByID(r.Context(), userID)
	if err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	respond.WriteJSON(w, http.StatusOK, resp)
}

// UpdateProfile обрабатывает PATCH /me/profile.
func (c *UserController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(context_model.ContextKeyUserID).(string)
	if !ok {
		respond.WriteProblem(w, r, c.log, error_model.New(error_model.KindUnauthorized, "user not authenticated"))
		return
	}

	var req request_model.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if err := c.svc.UpdateProfile(r.Context(), userID, req); err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
