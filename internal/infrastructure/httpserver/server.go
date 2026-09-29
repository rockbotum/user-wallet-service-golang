// Package httpserver собирает и настраивает HTTP-сервер.
package httpserver

import (
	"fmt"
	"net/http"
	"time"
)

const defaultReadHeaderTimeout = 5 * time.Second

// Config — настройки HTTP-сервера.
type Config struct {
	Port              int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
}

// New создаёт настроенный HTTP-сервер.
func New(handler http.Handler, cfg Config) *http.Server {
	headerTimeout := cfg.ReadHeaderTimeout
	if headerTimeout == 0 {
		headerTimeout = defaultReadHeaderTimeout
	}

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: headerTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}
}
