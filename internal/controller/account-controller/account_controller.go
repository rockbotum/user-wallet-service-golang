package account_controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"user-wallet-service/internal/controller/respond"
	context_model "user-wallet-service/internal/model/context-model"
	error_model "user-wallet-service/internal/model/error-model"
	request_model "user-wallet-service/internal/model/request-model"
	account_service "user-wallet-service/internal/service/account-service"
)

// AccountController обрабатывает HTTP-запросы со счетами.
type AccountController struct {
	svc *account_service.AccountService
	log *slog.Logger
}

// NewAccountController создаёт AccountController.
func NewAccountController(svc *account_service.AccountService, log *slog.Logger) *AccountController {
	return &AccountController{svc: svc, log: log}
}

// GetBalance обрабатывает GET /account.
func (c *AccountController) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(context_model.ContextKeyUserID).(string)
	if !ok {
		respond.WriteProblem(w, r, c.log, error_model.New(error_model.KindUnauthorized, "user not authenticated"))
		return
	}

	resp, err := c.svc.GetBalance(r.Context(), userID)
	if err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	respond.WriteJSON(w, http.StatusOK, resp)
}

// Deposit обрабатывает POST /account/deposit.
func (c *AccountController) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(context_model.ContextKeyUserID).(string)
	if !ok {
		respond.WriteProblem(w, r, c.log, error_model.New(error_model.KindUnauthorized, "user not authenticated"))
		return
	}

	var req request_model.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if err := validateAmount(req.Amount); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := c.svc.Deposit(r.Context(), userID, req.Amount); err != nil {
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Withdraw обрабатывает POST /account/withdraw.
func (c *AccountController) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(context_model.ContextKeyUserID).(string)
	if !ok {
		respond.WriteProblem(w, r, c.log, error_model.New(error_model.KindUnauthorized, "user not authenticated"))
		return
	}

	var req request_model.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	if err := validateAmount(req.Amount); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := c.svc.Withdraw(r.Context(), userID, req.Amount); err != nil {
		if error_model.Is(err, error_model.KindInvalid) {
			respond.WriteProblemInsufficient(w, r)
			return
		}
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Transfer обрабатывает POST /account/transfer.
func (c *AccountController) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(context_model.ContextKeyUserID).(string)
	if !ok {
		respond.WriteProblem(w, r, c.log, error_model.New(error_model.KindUnauthorized, "user not authenticated"))
		return
	}

	var req request_model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Bad request", map[string]string{"body": "invalid JSON"})
		return
	}

	errs := make(map[string]string)
	if req.ToUserID == "" {
		errs["to_user_id"] = "required"
	}
	if amountErr := validateAmount(req.Amount); amountErr != nil {
		errs["amount"] = amountErr["amount"]
	}
	if len(errs) > 0 {
		respond.WriteProblemWithErrors(w, r, http.StatusBadRequest, "Validation failed", errs)
		return
	}

	if err := c.svc.Transfer(r.Context(), userID, req.ToUserID, req.Amount); err != nil {
		if error_model.Is(err, error_model.KindInvalid) {
			respond.WriteProblemInsufficient(w, r)
			return
		}
		respond.WriteProblem(w, r, c.log, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// validateAmount проверяет, что сумма — положительное десятичное число с не более чем 2 знаками.
func validateAmount(amount string) map[string]string {
	if amount == "" {
		return map[string]string{"amount": "required"}
	}
	// простая проверка формата: digits, опционально точка/запятая + до 2 цифр
	valid := false
	dotFound := false
	digitsAfterDot := 0
	for i, ch := range amount {
		if ch >= '0' && ch <= '9' {
			if dotFound {
				digitsAfterDot++
				if digitsAfterDot > 2 {
					return map[string]string{"amount": "max 2 decimal places"}
				}
			}
			continue
		}
		if (ch == '.' || ch == ',') && i > 0 && !dotFound {
			dotFound = true
			continue
		}
		return map[string]string{"amount": "invalid format"}
	}
	if !valid && len(amount) > 0 {
		valid = true
	}
	if !valid {
		return map[string]string{"amount": "must be positive"}
	}
	return nil
}
