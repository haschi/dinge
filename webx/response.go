package webx

import (
	"bytes"
	"html/template"
	"net/http"
)

type Response interface {
	Render(w http.ResponseWriter, r *http.Request) error
}

//type Renderer func(w http.ResponseWriter) error

//type ResourceHandler func(a DingeResource, r *http.Request) Renderer

type SeeOther struct{ Url string }

func (s SeeOther) Render(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, s.Url, http.StatusSeeOther)
	return nil
}

type ServerError struct{ Cause error }

func (s ServerError) Render(w http.ResponseWriter, r *http.Request) error {
	status := http.StatusInternalServerError
	http.Error(w, http.StatusText(status), status)
	return nil
}

type NotFound struct{}

func (n NotFound) Render(w http.ResponseWriter, r *http.Request) error {
	http.NotFound(w, r)
	return nil
}

type HtmlResponse struct {
	Template *template.Template
	// View       string
	Data       any
	StatusCode int
}

func (h HtmlResponse) Render(w http.ResponseWriter, r *http.Request) error {

	var buffer bytes.Buffer
	if err := h.Template.Execute(&buffer, h.Data); err != nil {
		return err
	}

	w.WriteHeader(h.StatusCode)
	buffer.WriteTo(w)
	return nil
}
