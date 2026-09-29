package health

import (
	"context"
	"net/http"
	"time"

	"user-wallet-service/internal/controller/respond"

	"github.com/jmoiron/sqlx"
)

// HealthController отдаёт статус сервиса.
type HealthController struct {
	db *sqlx.DB
}

// NewHealthController создаёт обработчик проверки живости.
func NewHealthController(db *sqlx.DB) *HealthController {
	return &HealthController{db: db}
}

// Check обрабатывает GET /health.
func (h *HealthController) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	details := map[string]string{}

	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			status = "degraded"
			details["database"] = "unhealthy"
		} else {
			details["database"] = "healthy"
		}
	}

	httpStatus := http.StatusOK
	if status == "degraded" {
		httpStatus = http.StatusServiceUnavailable
	}

	respond.WriteJSON(w, httpStatus, map[string]any{
		"status":  status,
		"details": details,
	})
}
