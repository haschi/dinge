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

// Was hier fehlt ist die Middleware Kette, so dass Fehler ggf. von
// dem umschließenden Handler geloggt werden können.
func (a DingeApplication) HandleGet(w http.ResponseWriter, r *http.Request) {

	// Kommt nach DingeApplication
	var index, err = template.ParseFS(
		TemplatesFileSystem,
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

func (a DingeApplication) HandlePost(w http.ResponseWriter, r *http.Request) {

	var index, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/index/*.tmpl")

	if err != nil {
		a.Error(w, r, err)
		return
	}

	form := validation.Form{Request: r}

	var code string
	var anzahl int

	err = form.Scan(
		validation.Field(Code, validation.String(&code), validation.IsNotBlank),
		validation.Field(Anzahl, validation.Integer(&anzahl), validation.Min(1)),
	)

	if err != nil {
		a.Error(w, r, err)
		return
	}

	if !form.IsValid() {
		// a.Error(w, r, fmt.Errorf("fehler in den übermittelten Daten"))

		dinge, err := a.Repository.GetLatest()
		if err != nil {
			a.Error(w, r, err)
			return
		}

		data := Data{
			LetzteEinträge: dinge,
			Form: Form{
				Code:   code,
				Anzahl: anzahl,
			},
			ValidationErrors: form.ValidationErrors,
		}

		render(w, http.StatusUnprocessableEntity, index, data)
		return
	}

	a.Logger.Info("got post form",
		slog.String(string(Code), code),
		slog.Int(string(Anzahl), anzahl))

	id, err := a.Repository.Insert(r.Context(), code, anzahl)
	if err != nil {
		a.Error(w, r, err)
		return
	}

	a.Logger.Info("insert database record", slog.Int64("id", id))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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

func (a DingeApplication) HandleGetDing(w http.ResponseWriter, r *http.Request) {

	var page, err = template.ParseFS(
		TemplatesFileSystem,
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
		Id: id,
		Form: PostDingForm{
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
		TemplatesFileSystem,
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

	form := validation.Form{Request: r}

	var result PostDingForm

	err = form.Scan(
		validation.Field(Name, validation.String(&result.Name), validation.IsNotBlank),
		validation.Field(Code, validation.String(&result.Code), validation.IsNotBlank),
		validation.Field(Anzahl, validation.Integer(&result.Anzahl), validation.Min(1)),
	)

	if err != nil {
		a.Error(w, r, err)
		return
	}

	if !form.IsValid() {
		data := PostDingData{
			Id:               id,
			Form:             result,
			ValidationErrors: form.ValidationErrors,
		}

		if err := render(w, http.StatusOK, page, data); err != nil {
			a.Error(w, r, err)
			return
		}
	}

	err = a.Repository.NamenAktualisieren(id, result.Name)
	if err != nil {
		a.Error(w, r, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%v", id), http.StatusSeeOther)
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

func (a DingeApplication) handleGetEntnehmenCode(w http.ResponseWriter, r *http.Request) {
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
			a.Error(w, r, err)
		}

		return
	}

	a.Logger.Info("Ding gefunden", slog.String("code", code), slog.Any("ding", ding))

	url := fmt.Sprintf("/entnehmen/%v/menge", ding.Id)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (a DingeApplication) handleGetEntnehmenMenge(w http.ResponseWriter, r *http.Request) {

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
		a.Error(w, r, err)
		return
	}

	if err := render(w, http.StatusOK, page, data); err != nil {
		a.Error(w, r, err)
		return
	}

}

func (a DingeApplication) handlePostEntnehmen(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		a.Error(w, r, err)
		return
	}

	form := validation.Form{Request: r}

	var menge int

	err = form.Scan(
		validation.Field("menge", validation.Integer(&menge), validation.Min(1)),
	)

	if err != nil {
		a.Error(w, r, err)
		return
	}

	if !form.IsValid() {
		page, err := template.ParseFS(
			TemplatesFileSystem,
			"templates/layout/*.tmpl",
			"templates/pages/entnehmen/menge/*.tmpl")

		if err != nil {
			a.Error(w, r, err)
			return
		}

		ding, err := a.Repository.GetById(id)
		if err != nil {
			a.Error(w, r, err)
			return
		}
		data := struct {
			Ding             model.Ding
			Menge            int
			ValidationErrors validation.ErrorMap
		}{
			Ding:             ding,
			Menge:            1,
			ValidationErrors: form.ValidationErrors,
		}

		if err := render(w, http.StatusOK, page, data); err != nil {
			a.Error(w, r, err)
			return
		}
	}

	err = a.Repository.MengeAktualisieren(r.Context(), id, -menge)
	if err != nil {
		a.Error(w, r, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%v", id), http.StatusSeeOther)
}
