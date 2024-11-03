package main

import "net/http"

func routes(dinge DingeResource) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", newStaticHandler(dinge.Logger))
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

	mux.HandleFunc("GET /{$}", redirectTo("/dinge"))                         // Redirect to /dinge
	mux.HandleFunc("GET /dinge", ResponseWrapper(dinge.Logger, dinge.Index)) // Redirect to /dinge
	mux.HandleFunc("GET /dinge/new", ResponseWrapper(dinge.Logger, dinge.NewForm))
	mux.HandleFunc("POST /dinge", ResponseWrapper(dinge.Logger, dinge.Create))
	mux.HandleFunc("GET /dinge/{id}", ResponseWrapper(dinge.Logger, dinge.Show))
	mux.HandleFunc("GET /dinge/{id}/edit", ResponseWrapper(dinge.Logger, dinge.Edit))
	mux.HandleFunc("POST /dinge/{id}", ResponseWrapper(dinge.Logger, dinge.Update)) // Update

	mux.HandleFunc("GET /entnehmen", ResponseWrapper(dinge.Logger, handleGetEntnehmen))
	mux.HandleFunc("POST /entnehmen/{id}", ResponseWrapper(dinge.Logger, dinge.Destroy))
	mux.HandleFunc("GET /entnehmen/code", ResponseWrapper(dinge.Logger, dinge.handleGetEntnehmenCode)) // Destroy (Referenzzählung) => GET /:id Show aber mit Code statt Id

	mux.HandleFunc("GET /entnehmen/{id}/menge", ResponseWrapper(dinge.Logger, dinge.handleGetEntnehmenMenge)) // Liefert eine Form für die Entnahme

	mux.HandleFunc("GET /about", ResponseWrapper(dinge.Logger, handleAbout))
	return mux
}
