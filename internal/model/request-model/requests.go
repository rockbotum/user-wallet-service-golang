package request_model

import "time"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ProfileUpdateRequest struct {
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	BirthDate *time.Time `json:"birth_date"`
}

// Суммы принимаются строками: точное десятичное значение без артефактов float,
// валидация формата (число, не меньше 0, не более 2 знаков после запятой) — на границе.

type DepositRequest struct {
	Amount string `json:"amount"`
}

type WithdrawRequest struct {
	Amount string `json:"amount"`
}

type TransferRequest struct {
	ToUserID string `json:"to_user_id"`
	Amount   string `json:"amount"`
}

type TransactionFilter struct {
	Type      *string `json:"type"`
	Status    *string `json:"status"`
	DateFrom  *string `json:"date_from"`
	DateTo    *string `json:"date_to"`
	AmountMin *string `json:"amount_min"`
	AmountMax *string `json:"amount_max"`
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
	Sort      string  `json:"sort"`
}
