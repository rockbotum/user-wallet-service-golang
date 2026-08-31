package context_model

type contextKey string

const (
	// ContextKeyRequestID — идентификатор запроса.
	ContextKeyRequestID contextKey = "request_id"
	// ContextKeyUserID — UUID текущего пользователя.
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyRoleID — ID роли текущего пользователя.
	ContextKeyRoleID contextKey = "role_id"
)
