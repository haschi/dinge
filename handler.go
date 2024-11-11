package main

import (
	"fmt"
	"html/template"
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

func redirectTo(route string) func(*http.Request) webx.Response {
	return func(r *http.Request) webx.Response {
		return webx.PermanentRedirect(route)
	}
}

func handleAbout(r *http.Request) webx.Response {
	template, err := GetTemplate("about")
	if err != nil {
		return webx.ServerError(err)
	}

	return webx.HtmlResponse{Template: template, StatusCode: http.StatusOK}
}

// Zeigt eine Form an, um Dinge zu entnehmen.
func handleGetEntnehmen(r *http.Request) webx.Response {
	template, err := GetTemplate("entnehmen")
	if err != nil {
		return webx.ServerError(err)
	}

	return webx.HtmlResponse{Template: template, StatusCode: http.StatusOK}
}

func (a DingeResource) handleGetEntnehmenCode(r *http.Request) webx.Response {

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
			return webx.ServerError(err)
		}

		return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusNotFound}
	}

	return webx.SeeOther("/entnehmen/%v/menge", ding.Id)
}

// Liefert eine Form für ein spezifisches Ding, in der die Anzahl zu entfernender Exemplarer des Dings eingegeben werden kann. Die Anfrage wird dann an /entnehmen/:id gesendet.
func (a DingeResource) handleGetEntnehmenMenge(r *http.Request) webx.Response {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return webx.ServerError(err)
	}

	ding, err := a.Repository.GetById(id)
	if err != nil {
		return webx.ServerError(err)
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
		return webx.ServerError(err)
	}

	return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
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
