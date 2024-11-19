package webx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// SeeOther erzeugt eine Umleitung mit HTTP Status Code 303 See Other
func SeeOther(format string, args ...any) http.HandlerFunc {
	url := fmt.Sprintf(format, args...)
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

func PermanentRedirect(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, http.StatusPermanentRedirect)
	}
}

// ServerError erzeugt eine Antwort mit HTTP Status Code 500 Internal Server Error
//
// TODO: cause muss ausgegeben werden, aber nicht im Body, sondern im Protokoll.
func ServerError(w http.ResponseWriter, cause error) {
	status := http.StatusInternalServerError
	http.Error(w, http.StatusText(status), status)
	// TODO Log
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
func (h HtmlResponse) Render(w http.ResponseWriter) error {

	var buffer bytes.Buffer
	if err := h.Template.Execute(&buffer, h.Data); err != nil {
		return err
	}

	w.WriteHeader(h.StatusCode)
	if _, err := buffer.WriteTo(w); err != nil {
		return err
	}
	return nil
}
