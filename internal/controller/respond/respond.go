// Package respond содержит HTTP-обработчики.
package respond

import (
	"encoding/json"
	"log/slog"
	"net/http"

	error_model "user-wallet-service/internal/model/error-model"
	problem_model "user-wallet-service/internal/model/problem-model"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteProblem отправляет ошибку в формате RFC 9457 (application/problem+json).
func WriteProblem(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	var prob problem_model.ProblemDetail

	kind := extractKind(err)
	prob.Type = problem_model.KindProbType[kind]
	prob.Title = problem_model.KindTitles[kind]
	prob.Detail = err.Error()
	prob.Instance = r.URL.Path

	switch kind {
	case error_model.KindInvalid:
		prob.Status = http.StatusBadRequest
	case error_model.KindUnauthorized:
		prob.Status = http.StatusUnauthorized
	case error_model.KindForbidden:
		prob.Status = http.StatusForbidden
	case error_model.KindNotFound:
		prob.Status = http.StatusNotFound
	case error_model.KindConflict:
		prob.Status = http.StatusConflict
	case error_model.KindRateLimit:
		prob.Status = http.StatusTooManyRequests
	default:
		prob.Status = http.StatusInternalServerError
		prob.Detail = "internal error"
		log.Error("internal error", "error", err)
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(prob.Status)
	_ = json.NewEncoder(w).Encode(prob)
}

// WriteProblemWithErrors отправляет ошибку валидации с картой ошибок по полям.
func WriteProblemWithErrors(w http.ResponseWriter, r *http.Request, status int, title string, errs map[string]string) {
	prob := problem_model.ProblemDetail{
		Type:     problem_model.ProbTypeInvalid,
		Title:    title,
		Status:   status,
		Detail:   "validation failed",
		Instance: r.URL.Path,
		Errors:   errs,
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(prob.Status)
	_ = json.NewEncoder(w).Encode(prob)
}

// WriteProblemInsufficient отправляет ошибку недостатка средств (422).
func WriteProblemInsufficient(w http.ResponseWriter, r *http.Request) {
	prob := problem_model.ProblemDetail{
		Type:     problem_model.ProbTypeInsufficient,
		Title:    "Insufficient funds",
		Status:   http.StatusUnprocessableEntity,
		Detail:   "account balance is lower than requested amount",
		Instance: r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(prob.Status)
	_ = json.NewEncoder(w).Encode(prob)
}

// extractKind извлекает Kind из доменной ошибки. Если это не model.Error — возвращает KindInternal.
func extractKind(err error) error_model.Kind {
	if error_model.Is(err, error_model.KindInvalid) {
		return error_model.KindInvalid
	}
	if error_model.Is(err, error_model.KindUnauthorized) {
		return error_model.KindUnauthorized
	}
	if error_model.Is(err, error_model.KindForbidden) {
		return error_model.KindForbidden
	}
	if error_model.Is(err, error_model.KindNotFound) {
		return error_model.KindNotFound
	}
	if error_model.Is(err, error_model.KindConflict) {
		return error_model.KindConflict
	}
	if error_model.Is(err, error_model.KindRateLimit) {
		return error_model.KindRateLimit
	}
	return error_model.KindInternal
}
