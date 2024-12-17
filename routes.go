package main

import (
	"log/slog"
	"net/http"

	"github.com/haschi/dinge/webx"
)

// TODO: Bessere Namen für combine und compose wählen
// TODO: Verschieben nach webx.
func compose(m1, m2 webx.Middleware) webx.Middleware {
	return webx.Middleware(func(next http.Handler) http.Handler {
		return m1(m2(next))
	})
}

func routes(logger *slog.Logger, dinge DingeResource) http.Handler {
	mux := http.NewServeMux()

	// middleware
	requestLogger := webx.LogRequest(logger)
	nostore := webx.NoStore(logger)
	defaultMiddleware := compose(requestLogger, nostore)

	// nostore vorerst beibehalten bis gh#26
	mux.Handle("GET /static/", webx.Combine(newStaticHandler(logger), nostore, requestLogger))

	// Es gibt noch kein favicon. Daher NotFound
	mux.Handle("GET /favicon.ico", http.NotFoundHandler())

	// Analog zu RoR:
	//
	// GET  /dinge          dinge#index   Zeige eine Liste aller Dinge
	// GET  /dinge/new      dinge#new     Liefert eine HTML Form um ein neues Ding zu erzeugen
	// POST /dinge          dinge#create  Erzeugt eine neues Ding
	// GET  /dinge/:id      dinge#show    Zeigt ein spezifisches Ding an
	// GET  /dinge/:id/edit dinge#edit    Liefert eine HTML Form um ein spezifisches Ding zu bearbeiten
	// PATCH/PUT /dinge/:id dinge#update  Aktualisiert ein spezifisches Ding
	// DELETE /dinge/:id    dinge#destroy Löscht ein spezfisches Ding

	mux.Handle("GET /{$}", webx.Combine(http.RedirectHandler("/dinge", http.StatusPermanentRedirect), requestLogger))
	mux.Handle("GET /dinge", webx.CombineFunc(dinge.Index, requestLogger, nostore))
	mux.Handle("GET /dinge/new", webx.CombineFunc(dinge.NewForm, defaultMiddleware))
	mux.Handle("POST /dinge", webx.CombineFunc(dinge.Create, requestLogger))
	mux.Handle("GET /dinge/{id}", webx.CombineFunc(dinge.Show, defaultMiddleware))
	mux.Handle("GET /dinge/{id}/edit", webx.CombineFunc(dinge.Edit, defaultMiddleware))
	mux.Handle("POST /dinge/{id}", webx.CombineFunc(dinge.Update, requestLogger)) // Update
	mux.Handle("GET /dinge/{id}/photo", webx.CombineFunc(dinge.PhotoForm, defaultMiddleware))
	mux.Handle("POST /dinge/{id}/photo", webx.CombineFunc(dinge.PhotoUpload, requestLogger))
	mux.Handle("GET /photos/{id}", webx.CombineFunc(dinge.PhotoDownload, requestLogger))
	mux.Handle("GET /dinge/delete", webx.CombineFunc(dinge.DestroyForm, defaultMiddleware))
	mux.Handle("POST /dinge/delete", webx.CombineFunc(dinge.Destroy, requestLogger))

	mux.Handle("GET /about", webx.CombineFunc(handleAbout, requestLogger))
	return mux
}
