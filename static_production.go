//go:build !development

package main

import (
	"embed"
	"log/slog"
	"net/http"
)

//go:embed "static/css/*"
var Static embed.FS

func newStaticHandler(logger *slog.Logger) http.Handler {
	logger.Info("serving static assets from compiled virtual filesystem")
	return http.FileServerFS(Static)
}
