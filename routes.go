package main

import (
	"log/slog"
	"net/http"

	"github.com/haschi/dinge/webx"
)

func combine(handler http.HandlerFunc, middleware ...webx.Middleware) http.Handler {
	if len(middleware) == 0 {
		return handler
	}

	first := middleware[0]
	next := combine(handler, middleware[1:]...)
	return first(next)
}

// TODO: Bessere Namen für combine und compose wählen
// TODO: Verschieben nach webx.
func compose(m1, m2 webx.Middleware) webx.Middleware {
	return webx.Middleware(func(next http.Handler) http.Handler {
		return m1(m2(next))
	})
}

func routes(logger *slog.Logger, dinge DingeResource) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", newStaticHandler(logger))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Analog zu RoR:
	//
	// GET  /dinge          dinge#index   Zeige eine Liste aller Dinge
	// GET  /dinge/new      dinge#new     Liefert eine HTML Form um ein neues Ding zu erzeugen
	// POST /dinge          dinge#create  Erzeugt eine neues Ding
	// GET  /dinge/:id      dinge#show    Zeigt ein spezifisches Ding an
	// GET  /dinge/:id/edit dinge#edit    Liefert eine HTML Form um ein spezifisches Ding zu bearbeiten
	// PATCH/PUT /dinge/:id dinge#update  Aktualisiert ein spezifisches Ding
	// DELETE /dinge/:id    dinge#destroy Löscht ein spezfisches Ding

	weblogger := webx.LogRequest(logger)
	nostore := webx.NoStore(logger)
	defaultMiddleware := compose(weblogger, nostore)

	// noop := webx.Noop
	mux.Handle("GET /{$}", combine(webx.PermanentRedirect("/dinge"), weblogger)) // Redirect to /dinge
	mux.Handle("GET /dinge", combine(dinge.Index, weblogger, nostore))           // Redirect to /dinge
	mux.Handle("GET /dinge/new", combine(dinge.NewForm, defaultMiddleware))
	mux.Handle("POST /dinge", combine(dinge.Create, weblogger))
	mux.Handle("GET /dinge/{id}", combine(dinge.Show, defaultMiddleware))
	mux.Handle("GET /dinge/{id}/edit", combine(dinge.Edit, defaultMiddleware))
	mux.Handle("POST /dinge/{id}", combine(dinge.Update, weblogger)) // Update

	mux.Handle("GET /dinge/delete", combine(DestroyForm, defaultMiddleware))
	mux.Handle("POST /dinge/delete", combine(dinge.Destroy, weblogger))

	mux.Handle("GET /about", combine(handleAbout, weblogger))
	return mux
}
