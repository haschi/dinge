package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/validation"
	"github.com/haschi/dinge/webx"
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

func redirectTo(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route, http.StatusPermanentRedirect)
	}
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	template, err := GetTemplate("about")
	if err != nil {
		webx.ServerError{Cause: err}.Render(w, r)
	}

	webx.HtmlResponse{Template: template, StatusCode: http.StatusOK}.Render(w, r)
}

// Zeigt eine Form an, um Dinge zu entnehmen.
func handleGetEntnehmen(w http.ResponseWriter, r *http.Request) {
	template, err := GetTemplate("entnehmen")
	if err != nil {
		webx.ServerError{Cause: err}.Render(w, r)
	}

	webx.HtmlResponse{Template: template, StatusCode: http.StatusOK}.Render(w, r)
}

func (a DingeResource) handleGetEntnehmenCode(w http.ResponseWriter, r *http.Request) {
	response := func() webx.Response {
		code := r.FormValue("code")

		ding, err := a.Repository.GetByCode(code)
		if err != nil {

			var data struct {
				ValidationErrors validation.ErrorMap
			}

			data.ValidationErrors = map[string]string{}
			data.ValidationErrors["code"] = "Unbekannter Produktcode"

			template, err := GetTemplate("entnehmen")
			if err != nil {
				return webx.ServerError{Cause: err}
			}

			return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusNotFound}
		}

		a.Logger.Info("Ding gefunden", slog.String("code", code), slog.Any("ding", ding))
		return webx.SeeOther{Url: fmt.Sprintf("/entnehmen/%v/menge", ding.Id)}
	}()

	response.Render(w, r)
}

// Liefert eine Form für ein spezifisches Ding, in der die Anzahl zu entfernender Exemplarer des Dings eingegeben werden kann. Die Anfrage wird dann an /entnehmen/:id gesendet.
func (a DingeResource) handleGetEntnehmenMenge(w http.ResponseWriter, r *http.Request) {

	response := func() webx.Response {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			return webx.ServerError{Cause: err}
		}

		ding, err := a.Repository.GetById(id)
		if err != nil {
			return webx.ServerError{Cause: err}
		}

		data := struct {
			Ding             model.Ding
			Menge            int
			ValidationErrors validation.ErrorMap
		}{
			Ding:  ding,
			Menge: 1,
		}

		template, err := GetTemplate("entnehmen/menge")
		if err != nil {
			return webx.ServerError{Cause: err}
		}

		return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
	}()

	response.Render(w, r)
}

func GetTemplate(name string) (*template.Template, error) {
	t, err := template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		fmt.Sprintf("templates/pages/%v/*.tmpl", name))

	if err != nil {
		return nil, err
	}

	return t, nil
}
