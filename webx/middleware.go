package webx

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Middleware ist HTTP Handler, der den Aufruf an seinen Nachfolger weiterleitet.
type Middleware func(next http.Handler) http.Handler

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// LogRequest ist ein Middleware Handler, der HTTP Anfragen protokolliert
func LogRequest(logger *slog.Logger) Middleware {

	handlerFunc := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			start := time.Now()

			next.ServeHTTP(wrapper, r)

			message := fmt.Sprintf("%v %v", r.Method, r.URL.Path)
			duration := time.Since(start)
			logger.Info(message,
				slog.Int("status", wrapper.statusCode),
				slog.Duration("duration", duration),
			)
		}
	}

	return func(next http.Handler) http.Handler {
		return handlerFunc(next.ServeHTTP)
	}
}

// Noop ist ein Middleware Handler ohne Funktion
func Noop(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// NoStore ist ein Middleware Handler, der Browser anweist die Antwort nicht zu Speichern.
//
// TODO: Nachforschen, ob das reicht. Quelle: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
func NoStore(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Set Cache-Control header")
			w.Header().Add("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
}
