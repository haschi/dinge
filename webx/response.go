package webx

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/haschi/dinge/validation"
)

// SeeOther erzeugt eine Umleitung mit HTTP Status Code 303 See Other
func SeeOther(format string, args ...any) http.HandlerFunc {
	url := fmt.Sprintf(format, args...)
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

// ServerError erzeugt eine Antwort mit HTTP Status Code 500 Internal Server Error
//
// TODO: cause muss ausgegeben werden, aber nicht im Body, sondern im Protokoll.
func ServerError(w http.ResponseWriter, cause error) {
	println(cause.Error())
	slog.Error(cause.Error())
	status := http.StatusInternalServerError
	http.Error(w, http.StatusText(status), status)
}

type TemplateData[T any] struct {
	Styles           []string
	Scripts          []string
	FormValues       T
	ValidationErrors validation.ErrorMap
}

// HtmlResponse liefert den HTML Inhalt für die Anwort auf eine Anfrage
type HtmlResponse[T any] struct {
	TemplateName string
	Data         TemplateData[T]
	StatusCode   int
}

// TODO Offen:
// 1. Ausgabe des Fehlers im Log
// 2. Doppelten Code vermeiden
func (h HtmlResponse[T]) Render(w http.ResponseWriter, fs fs.FS) error {

	var buffer bytes.Buffer
	template, err := getTemplate(fs, h.TemplateName)
	if err != nil {
		return err
	}

	if err := template.ExecuteTemplate(&buffer, "layout.tmpl", h.Data); err != nil {
		return err
	}

	w.WriteHeader(h.StatusCode)
	if _, err := buffer.WriteTo(w); err != nil {
		return err
	}
	return nil
}

func getTemplate(fs fs.FS, name string) (*template.Template, error) {

	t, err := template.New("").ParseFS(
		fs,
		fmt.Sprintf("templates/pages/%v/*.tmpl", name),
		"templates/common/*.tmpl",
		"templates/layout/*.tmpl",
	)

	if err != nil {
		return nil, err
	}

	return t, nil
}
