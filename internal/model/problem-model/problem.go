package problem_model

import error_model "user-wallet-service/internal/model/error-model"

// ProblemDetail описывает ошибку в формате RFC 9457 (Problem Details for HTTP APIs).
type ProblemDetail struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Status   int               `json:"status"`
	Detail   string            `json:"detail"`
	Instance string            `json:"instance,omitempty"`
	Errors   map[string]string `json:"errors,omitempty"`
}

// URI-идентификаторы типов ошибок для клиентов.
const (
	ProbTypeInvalid      = "https://example.com/problems/invalid-input"
	ProbTypeUnauthorized = "https://example.com/problems/unauthorized"
	ProbTypeForbidden    = "https://example.com/problems/forbidden"
	ProbTypeNotFound     = "https://example.com/problems/not-found"
	ProbTypeConflict     = "https://example.com/problems/conflict"
	ProbTypeInternal     = "https://example.com/problems/internal-error"
	ProbTypeRateLimit    = "https://example.com/problems/rate-limit-exceeded"
	ProbTypeInsufficient = "https://example.com/problems/insufficient-funds"
)

// KindTitles — краткие описания для каждого Kind.
var KindTitles = map[error_model.Kind]string{
	error_model.KindInvalid:      "Validation failed",
	error_model.KindUnauthorized: "Authentication required",
	error_model.KindForbidden:    "Access denied",
	error_model.KindNotFound:     "Resource not found",
	error_model.KindConflict:     "Resource conflict",
	error_model.KindInternal:     "Internal server error",
	error_model.KindRateLimit:    "Rate limit exceeded",
}

// KindProbType — маппинг Kind на URI-тип ошибки.
var KindProbType = map[error_model.Kind]string{
	error_model.KindInvalid:      ProbTypeInvalid,
	error_model.KindUnauthorized: ProbTypeUnauthorized,
	error_model.KindForbidden:    ProbTypeForbidden,
	error_model.KindNotFound:     ProbTypeNotFound,
	error_model.KindConflict:     ProbTypeConflict,
	error_model.KindInternal:     ProbTypeInternal,
	error_model.KindRateLimit:    ProbTypeRateLimit,
}
