package main

import (
	"log/slog"
	"net/http"

	"github.com/haschi/dinge/webx"
)

// TODO: Bessere Namen für combine und compose wählen
// TODO: Verschieben nach webx.
func compose(m1, m2 webx.MiddlewareFunc) webx.MiddlewareFunc {
	return webx.MiddlewareFunc(func(next http.Handler) http.Handler {
		return m1(m2(next))
	})
}

func routes(logger *slog.Logger, staticHandler http.Handler, aboutHandler webx.Module, dinge webx.Module, photos webx.Module) *http.ServeMux {
	mux := http.NewServeMux()

	// middleware
	requestLogger := webx.LogRequest(logger)
	nostore := webx.NoStore(logger)
	defaultMiddleware := webx.MiddlewareFunc(compose(requestLogger, nostore))

	mux.Handle("GET /static/", webx.Combine(staticHandler, requestLogger))

	// Es gibt noch kein favicon. Daher NotFound
	mux.Handle("GET /favicon.ico", webx.Combine(http.NotFoundHandler(), requestLogger))

	mux.Handle("GET /{$}", webx.Combine(
		http.RedirectHandler("/dinge/", http.StatusPermanentRedirect),
		requestLogger))

	dinge.Mount(mux, "/dinge", defaultMiddleware)
	aboutHandler.Mount(mux, "/about", defaultMiddleware)
	photos.Mount(mux, "/dinge/{id}", defaultMiddleware)

	return mux
}
