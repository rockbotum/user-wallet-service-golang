// Package logger предоставляет конструктор структурированного логгера.
package logger

import (
	"io"
	"log/slog"
)

// New создаёт JSON-логгер с указанным уровнем.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}
