package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Was hier fehlt ist die Middleware Kette, so dass Fehler ggf. von
// dem umschließenden Handler geloggt werden können.
func (a DingeApplication) HandleGet(w http.ResponseWriter, r *http.Request) {

	// Kommt nach DingeApplication
	var index, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/index/*.tmpl")

	if err != nil {
		a.Error(w, r, err)
		return
	}

	dinge, err := a.Repository.GetLatest()
	if err != nil {
		a.Error(w, r, err)
		return
	}

	data := Data{
		LetzteEinträge: dinge,
		Form:           Form{Anzahl: 1},
	}

	if err := render(w, http.StatusOK, index, data); err != nil {
		a.Error(w, r, err)
		return
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func PositiveInteger(value int) bool {
	return value >= 0
}

func formIsValid(errors map[string]string) bool {
	return len(errors) == 0
}

func FormGetStringValidate(form *url.Values, errors map[string]string, field string, message string, validator func(string) bool) string {
	var value = form.Get(field)
	if !validator(value) {
		errors[field] = message
		return ""
	}
	return value
}

func FormGetIntValidate(form *url.Values, errors map[string]string, field string, message string, validator func(int) bool) int {
	var str = form.Get(field)

	value, err := strconv.Atoi(str)
	if err != nil || !validator(value) {
		errors[field] = message
		return 0
	}

	return value
}

func (a DingeApplication) HandlePost(w http.ResponseWriter, r *http.Request) {

	var index, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/index/*.tmpl")

	if err != nil {
		a.Error(w, r, err)
	}

	err = r.ParseForm()
	if err != nil {
		a.Error(w, r, err)
		return
	}

	var validationError = make(map[string]string)

	code := FormGetStringValidate(&r.PostForm, validationError, "code", "Das Feld darf nicht leer sein", NotBlank)

	anzahl := FormGetIntValidate(&r.PostForm, validationError, "anzahl", "Das Feld muss eine Zahl größer 0 enthalten", PositiveInteger)

	if !formIsValid(validationError) {
		// a.Error(w, r, fmt.Errorf("fehler in den übermittelten Daten"))

		dinge, err := a.Repository.GetLatest()
		if err != nil {
			a.Error(w, r, err)
			return
		}

		data := Data{
			LetzteEinträge: dinge,
			Form: Form{
				Code:   r.PostForm.Get("code"),
				Anzahl: anzahl,
			},
			FieldErrors: validationError,
		}

		render(w, http.StatusUnprocessableEntity, index, data)
		return
	}

	a.Logger.Info("got post form",
		slog.String("code", code),
		slog.Int("anzahl", anzahl))

	id, err := a.Repository.Insert(r.Context(), code, anzahl)
	if err != nil {
		a.Error(w, r, err)
		return
	}

	a.Logger.Info("insert database record", slog.Int64("id", id))

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
}

type PostDingForm struct {
	Id     int64
	Name   string
	Code   string
	Anzahl int
}

type PostDingData struct {
	Form   PostDingForm
	Errors map[string]string
}

func (a DingeApplication) HandleGetDing(w http.ResponseWriter, r *http.Request) {

	var page, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/ding/*.tmpl")

	if err != nil {
		a.Error(w, r, err)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		a.Error(w, r, err)
		return
	}

	ding, err := a.Repository.GetById(id)
	if err != nil {
		a.Error(w, r, err)
		return
	}

	data := PostDingData{
		Form: PostDingForm{
			Id:     id,
			Name:   ding.Name,
			Code:   ding.Code,
			Anzahl: ding.Anzahl,
		},
	}

	if err := render(w, http.StatusOK, page, data); err != nil {
		a.Error(w, r, err)
		return
	}
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

func (a DingeApplication) HandlePostDing(w http.ResponseWriter, r *http.Request) {
	var page, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/ding/*.tmpl")

	if err != nil {
		a.Error(w, r, err)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		a.Error(w, r, err)
		return
	}

	validationErrors := make(map[string]string)
	name := FormGetStringValidate(&r.PostForm, validationErrors, "name", "Das Feld darf nicht leer sein", NotBlank)

	if !formIsValid(validationErrors) {
		data := PostDingData{
			Form: PostDingForm{
				Id:     id,
				Name:   name,
				Code:   "",
				Anzahl: 0,
			},
			Errors: validationErrors,
		}

		if err := render(w, http.StatusOK, page, data); err != nil {
			a.Error(w, r, err)
			return
		}
	}

	err = a.Repository.NamenAktualisieren(id, name)
	if err != nil {
		a.Error(w, r, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%v", id), http.StatusSeeOther)
}
