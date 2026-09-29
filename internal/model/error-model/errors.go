// Package error_model содержит доменные ошибки.
package error_model

import (
	"errors"
)

// Kind классифицирует доменную ошибку для стабильного маппинга на границах системы.
type Kind string

const (
	// KindInvalid — некорректные входные данные.
	KindInvalid Kind = "invalid"
	// KindNotFound — ресурс не найден.
	KindNotFound Kind = "not_found"
	// KindConflict — конфликт состояния.
	KindConflict Kind = "conflict"
	// KindUnauthorized — требуется аутентификация.
	KindUnauthorized Kind = "unauthorized"
	// KindForbidden — доступ запрещён.
	KindForbidden Kind = "forbidden"
	// KindInternal — внутренняя ошибка.
	KindInternal Kind = "internal"
	// KindRateLimit — превышен лимит запросов.
	KindRateLimit Kind = "rate_limit"
)

// Error — доменная ошибка с классификацией.
type Error struct {
	Kind    Kind
	Message string
	Err     error
}

// Error возвращает сообщение ошибки.
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Kind)
}

// Unwrap возвращает исходную ошибку.
func (e *Error) Unwrap() error {
	return e.Err
}

// New создаёт доменную ошибку с классификацией.
func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

// Wrap оборачивает исходную ошибку с классификацией и контекстом.
func Wrap(kind Kind, message string, err error) *Error {
	return &Error{Kind: kind, Message: message, Err: err}
}

// Is проверяет, относится ли err к указанной классификации.
func Is(err error, kind Kind) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind == kind
	}
	return false
}
