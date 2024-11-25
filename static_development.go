//go:build development

package main

import (
	"log/slog"
	"net/http"
)

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func newStaticHandler(logger *slog.Logger) http.Handler {
	logger.Info("serving static assets from disk")
	return noStore(http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))
}
