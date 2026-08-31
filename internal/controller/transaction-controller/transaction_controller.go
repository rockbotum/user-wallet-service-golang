package transaction_controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"user-wallet-service/internal/controller/respond"
	context_model "user-wallet-service/internal/model/context-model"
	error_model "user-wallet-service/internal/model/error-model"
	request_model "user-wallet-service/internal/model/request-model"
	transaction_service "user-wallet-service/internal/service/transaction-service"
)

// TransactionController обрабатывает HTTP-запросы истории транзакций.
type TransactionController struct {
	svc *transaction_service.TransactionService
	log *slog.Logger
}

// NewTransactionController создаёт TransactionController.
func NewTransactionController(svc *transaction_service.TransactionService, log *slog.Logger) *TransactionController {
	return &TransactionController{svc: svc, log: log}
}

// List обрабатывает GET /transactions.
func (c *TransactionController) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(context_model.ContextKeyUserID).(string)
	if !ok {
		respond.WriteProblem(w, r, c.log, error_model.New(error_model.KindUnauthorized, "user not authenticated"))
		return
	}

	filter := parseTransactionFilter(r)

	resp, err := c.svc.ListByAccount(r.Context(), userID, filter)
	if err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	respond.WriteJSON(w, http.StatusOK, resp)
}

// parseTransactionFilter извлекает параметры фильтрации из query string.
func parseTransactionFilter(r *http.Request) request_model.TransactionFilter {
	q := r.URL.Query()
	filter := request_model.TransactionFilter{
		Page:  defaultInt(q.Get("page"), 1),
		Limit: defaultInt(q.Get("limit"), 20),
		Sort:  q.Get("sort"),
	}

	if v := q.Get("type"); v != "" {
		filter.Type = &v
	}
	if v := q.Get("status"); v != "" {
		filter.Status = &v
	}
	if v := q.Get("date_from"); v != "" {
		filter.DateFrom = &v
	}
	if v := q.Get("date_to"); v != "" {
		filter.DateTo = &v
	}
	if v := q.Get("amount_min"); v != "" {
		filter.AmountMin = &v
	}
	if v := q.Get("amount_max"); v != "" {
		filter.AmountMax = &v
	}

	return filter
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
