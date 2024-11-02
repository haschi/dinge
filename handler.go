package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/validation"
)

const (
	Name   = "name"
	Anzahl = "anzahl"
	Code   = "code"
)

type PostDingForm struct {
	Name   string
	Code   string
	Anzahl int
}

type PostDingData struct {
	Id               int64
	Form             PostDingForm
	ValidationErrors validation.ErrorMap
}

func render(w http.ResponseWriter, status int, page *template.Template, data any) error {
	var buffer bytes.Buffer
	if err := page.Execute(&buffer, data); err != nil {
		return err
	}

	w.WriteHeader(status)
	buffer.WriteTo(w)
	return nil
}

func redirectTo(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route, http.StatusPermanentRedirect)
	}
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	var page, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/about/*.tmpl")

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	render(w, http.StatusOK, page, nil)
}

// Zeigt eine Form an, um Dinge zu entnehmen.
func handleGetEntnehmen(w http.ResponseWriter, r *http.Request) {
	var page, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/entnehmen/*.tmpl")

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	render(w, http.StatusOK, page, nil)
}

func (a DingeResource) handleGetEntnehmenCode(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")

	ding, err := a.Repository.GetByCode(code)
	if err != nil {

		var page, err = template.ParseFS(
			TemplatesFileSystem,
			"templates/layout/*.tmpl",
			"templates/pages/entnehmen/*.tmpl")

		var data struct {
			ValidationErrors validation.ErrorMap
		}

		data.ValidationErrors = map[string]string{}
		data.ValidationErrors["code"] = "Unbekannter Produktcode"

		if err = render(w, http.StatusNotFound, page, data); err != nil {
			a.ServerError(w, r, err)
		}

		return
	}

	a.Logger.Info("Ding gefunden", slog.String("code", code), slog.Any("ding", ding))

	url := fmt.Sprintf("/entnehmen/%v/menge", ding.Id)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// Liefert eine Form für ein spezifisches Ding, in der die Anzahl zu entfernender Exemplarer des Dings eingegeben werden kann. Die Anfrage wird dann an /entnehmen/:id gesendet.
func (a DingeResource) handleGetEntnehmenMenge(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		a.ServerError(w, r, err)
		return
	}

	ding, err := a.Repository.GetById(id)
	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	data := struct {
		Ding             model.Ding
		Menge            int
		ValidationErrors validation.ErrorMap
	}{
		Ding:  ding,
		Menge: 1,
	}

	page, err := template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/entnehmen/menge/*.tmpl")

	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	if err := render(w, http.StatusOK, page, data); err != nil {
		a.ServerError(w, r, err)
		return
	}

}
