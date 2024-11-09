package webx

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Middleware Function
func Log(logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		wrapper := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapper, r)

		logger.Info(fmt.Sprintf("%v %v", r.Method, r.URL.Path), slog.Int("status", wrapper.statusCode))
	}
}

type Response interface {
	Render(w http.ResponseWriter, r *http.Request)
}

type SeeOther struct{ Url string }

func (s SeeOther) Render(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.Url, http.StatusSeeOther)
}

type ServerError struct{ Cause error }

func (s ServerError) Render(w http.ResponseWriter, r *http.Request) {
	status := http.StatusInternalServerError
	http.Error(w, http.StatusText(status), status)
}

type NotFound struct{}

func (n NotFound) Render(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

type HtmlResponse struct {
	Template   *template.Template
	Data       any
	StatusCode int
}

func (h HtmlResponse) Render(w http.ResponseWriter, r *http.Request) {

	var buffer bytes.Buffer
	if err := h.Template.Execute(&buffer, h.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(h.StatusCode)
	buffer.WriteTo(w)
	// TODO: Fehlerbehandlung
}
