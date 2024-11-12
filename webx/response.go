package webx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// Response ist eine Antwort auf eine HTTP Anfrage
type Response interface {
	// Render gibt die Antwort in einem [http.ResponseWriter] aus.
	Render(w http.ResponseWriter)
}

// ResponseFunc ist eine Funktion, mit der die Antwort auf eine Anfrage ausgegeben wird.
//
// Implementiert aber die Response Schnittstelle
type ResponseFunc func(http.ResponseWriter)

// RenderFunc implementiert [Response.Render]
func (f ResponseFunc) Render(w http.ResponseWriter) {
	f(w)
}

// NotFound erzeugt eine Antwort auf eine Anfrage mit einem HTTP 404 Not Found Fehler.
func NotFound(r *http.Request) ResponseFunc {
	return func(w http.ResponseWriter) {
		http.NotFound(w, r)
	}
}

// SeeOther erzeugt eine Umleitung mit HTTP Status Code 303 See Other
func SeeOther(r *http.Request, format string, a ...any) ResponseFunc {
	url := fmt.Sprintf(format, a...)
	return func(w http.ResponseWriter) {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

// PermanentRedirect erzeugt eine Umleitung mit HTTP Status Code 308 Permanent Redirect
func PermanentRedirect(r *http.Request, route string) ResponseFunc {
	return func(w http.ResponseWriter) {
		http.Redirect(w, r, route, http.StatusPermanentRedirect)
	}
}

// ServerError erzeugt eine Antwort mit HTTP Status Code 500 Internal Server Error
//
// TODO: cause muss ausgegeben werden, aber nicht im Body, sondern im Protokoll.
func ServerError(cause error) ResponseFunc {
	return func(w http.ResponseWriter) {
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
	}
}

// HtmlResponse liefert den HTML Inhalt für die Anwort auf eine Anfrage
type HtmlResponse struct {
	Template   *template.Template
	Data       any
	StatusCode int
}

// TODO Offen:
// 1. Ausgabe des Fehlers im Log
// 2. Doppelten Code vermeiden
func (h HtmlResponse) Render(w http.ResponseWriter) {

	var buffer bytes.Buffer
	if err := h.Template.Execute(&buffer, h.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(h.StatusCode)
	buffer.WriteTo(w)
	// TODO: Fehlerbehandlung
}
