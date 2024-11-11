package webx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

type Response interface {
	// TODO: http.Request nicht als Parameter. Wenn der Request benötigt wird, diesen über eine Closure oder Struktur binden.
	Render(w http.ResponseWriter, r *http.Request)
}

// ResponseFunc ist ein Alias für HandlerFunc
//
// Implementiert aber die Response Schnittstelle
// TODO: Eigene Signatur ohne Request.
type ResponseFunc http.HandlerFunc

// RenderFunc implementiert Response
func (f ResponseFunc) Render(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

func NotFound() Response {
	return ResponseFunc(http.NotFound)
}

func SeeOther(format string, a ...any) Response {
	url := fmt.Sprintf(format, a...)
	return ResponseFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusSeeOther)
	})
}

func PermanentRedirect(route string) Response {
	return ResponseFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route, http.StatusPermanentRedirect)
	})
}

// TODO: cause muss ausgegeben werden, aber nicht im Body, sondern im Protokoll.
func ServerError(cause error) ResponseFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
	}
}

type HtmlResponse struct {
	Template   *template.Template
	Data       any
	StatusCode int
}

// TODO Offen:
// 1. Ausgabe des Fehlers im Log
// 2. Doppelten Code vermeiden
func (h HtmlResponse) Render(w http.ResponseWriter, r *http.Request) {

	var buffer bytes.Buffer
	if err := h.Template.Execute(&buffer, h.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(h.StatusCode)
	buffer.WriteTo(w)
	// TODO: Fehlerbehandlung
}
