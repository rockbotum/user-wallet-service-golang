package admin_controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"user-wallet-service/internal/controller/respond"
	admin_service "user-wallet-service/internal/service/admin-service"
)

// AdminController обрабатывает HTTP-запросы администраторского управления.
type AdminController struct {
	svc *admin_service.AdminService
	log *slog.Logger
}

// NewAdminController создаёт AdminController.
func NewAdminController(svc *admin_service.AdminService, log *slog.Logger) *AdminController {
	return &AdminController{svc: svc, log: log}
}

// ListUsers обрабатывает GET /admin/users.
func (c *AdminController) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := defaultInt(q.Get("page"), 1)
	limit := defaultInt(q.Get("limit"), 20)
	offset := (page - 1) * limit

	users, total, err := c.svc.ListUsers(r.Context(), offset, limit)
	if err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	respond.WriteJSON(w, http.StatusOK, map[string]any{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// BlockUser обрабатывает PATCH /admin/users/{id}/block.
func (c *AdminController) BlockUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", map[string]string{"id": "required"})
		return
	}

	if err := c.svc.BlockUser(r.Context(), id); err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UnblockUser обрабатывает PATCH /admin/users/{id}/unblock.
func (c *AdminController) UnblockUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", map[string]string{"id": "required"})
		return
	}

	if err := c.svc.UnblockUser(r.Context(), id); err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateRole обрабатывает PATCH /admin/users/{id}/role.
func (c *AdminController) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", map[string]string{"id": "required"})
		return
	}

	var req struct {
		RoleID int `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if req.RoleID <= 0 {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", map[string]string{"role_id": "must be positive"})
		return
	}

	if err := c.svc.UpdateRole(r.Context(), id, req.RoleID); err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListRoles обрабатывает GET /admin/roles.
func (c *AdminController) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.svc.ListRoles(r.Context())
	if err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	respond.WriteJSON(w, http.StatusOK, map[string]any{"roles": roles})
}

func defaultInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
