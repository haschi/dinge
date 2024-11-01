package main

import "net/http"

func routes(dinge DingeApplication) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", dinge.HandleGet)
	mux.HandleFunc("POST /{$}", dinge.HandlePost)
	mux.Handle("GET /static/", newStaticHandler(dinge.Logger))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /about", handleAbout)
	mux.HandleFunc("GET /entnehmen", handleGetEntnehmen)
	mux.HandleFunc("POST /entnehmen/{id}", dinge.handlePostEntnehmen)
	mux.HandleFunc("GET /entnehmen/code", dinge.handleGetEntnehmenCode)
	mux.HandleFunc("GET /entnehmen/{id}/menge", dinge.handleGetEntnehmenMenge)
	mux.HandleFunc("GET /{id}", dinge.HandleGetDing)
	mux.HandleFunc("POST /{id}", dinge.HandlePostDing)
	return mux
}
