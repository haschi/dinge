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

func handleAbout(r *http.Request) Renderer {
	return HtmlResponse("about", nil, http.StatusOK)
}

// Zeigt eine Form an, um Dinge zu entnehmen.
func handleGetEntnehmen(r *http.Request) Renderer {
	return HtmlResponse("entnehmen", nil, http.StatusOK)
}

func (a DingeResource) handleGetEntnehmenCode(r *http.Request) Renderer {
	code := r.FormValue("code")

	ding, err := a.Repository.GetByCode(code)
	if err != nil {

		var data struct {
			ValidationErrors validation.ErrorMap
		}

		data.ValidationErrors = map[string]string{}
		data.ValidationErrors["code"] = "Unbekannter Produktcode"

		return HtmlResponse("entnehmen", data, http.StatusNotFound)
	}

	a.Logger.Info("Ding gefunden", slog.String("code", code), slog.Any("ding", ding))
	return SeeOther(r, fmt.Sprintf("/entnehmen/%v/menge", ding.Id))
}

// Liefert eine Form für ein spezifisches Ding, in der die Anzahl zu entfernender Exemplarer des Dings eingegeben werden kann. Die Anfrage wird dann an /entnehmen/:id gesendet.
func (a DingeResource) handleGetEntnehmenMenge(r *http.Request) Renderer {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return ServerError(err)
	}

	ding, err := a.Repository.GetById(id)
	if err != nil {
		return ServerError(err)
	}

	data := struct {
		Ding             model.Ding
		Menge            int
		ValidationErrors validation.ErrorMap
	}{
		Ding:  ding,
		Menge: 1,
	}

	return HtmlResponse("entnehmen/menge", data, http.StatusOK)
}
