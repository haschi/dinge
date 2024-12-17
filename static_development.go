//go:build development

package main

import (
	"log/slog"
	"net/http"

	"github.com/haschi/dinge/webx"
)

func newStaticHandler(logger *slog.Logger) http.Handler {
	logger.Info("serving static assets from disk")
	return webx.Combine(
		http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))),
		webx.NoStore(logger),
	)
}
