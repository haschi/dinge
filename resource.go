package main

import (
	"net/http"
	"strconv"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/validation"
	"github.com/haschi/dinge/webx"
)

type DingeResource struct {
	Repository model.Repository
}

// Liefert eine HTML Form zum Erzeugen eines neuen Dings.
func (a DingeResource) NewForm(r *http.Request) webx.Response {

	data := FormData[CreateData]{
		Form: CreateData{Anzahl: 1},
	}

	template, err := GetTemplate("new")
	if err != nil {
		return webx.ServerError(err)
	}

	return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
}

// Zeigt eine Liste aller Dinge
func (a DingeResource) Index(r *http.Request) webx.Response {

	dinge, err := a.Repository.GetLatest()
	if err != nil {
		return webx.ServerError(err)
	}

	data := Data{
		LetzteEinträge: dinge,
		Form:           Form{Anzahl: 1},
	}

	template, err := GetTemplate("index")
	if err != nil {
		return webx.ServerError(err)
	}

	return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
}

type FormData[T any] struct {
	ValidationErrors validation.ErrorMap
	Form             T
}

type CreateData struct {
	Code   string
	Anzahl int
}

// Erzeugt ein neues Ding.
//
// Ziel der Form von NewForm.
//
// Wenn ein Fehler auftritt, wird die Form von [NewForm] mit den übertragenen
// Werte erneut angezeigt. Zusätzlich werden die Validierungsfehler ausgegeben.
//
// Wenn die Erzeugung eine neuen Dings erfolgreich war, wird nach /new
// weitergeleitet, wenn das Ding bekannt ist, so dass der Workflow fortgesetzt
// werden kann und weitere Dinge hinzugefügt werden können. Wenn es sich um ein
// neues Ding handelt, wird nach /:id/edit weitergeleitet, um weitere Daten über
// das Ding anzufordern.
func (a DingeResource) Create(r *http.Request) webx.Response {

	form := validation.Form{Request: r}

	var content CreateData

	err := form.Scan(
		validation.Field(Code, validation.String(&content.Code), validation.IsNotBlank),
		validation.Field(Anzahl, validation.Integer(&content.Anzahl), validation.Min(1)),
	)

	if err != nil {
		return webx.ServerError(err)
	}

	if !form.IsValid() {
		// "fehler in den übermittelten Daten"

		data := FormData[CreateData]{
			Form:             content,
			ValidationErrors: form.ValidationErrors,
		}

		template, err := GetTemplate("new")
		if err != nil {
			return webx.ServerError(err)
		}
		return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusUnprocessableEntity}
	}

	result, err := a.Repository.Insert(r.Context(), content.Code, content.Anzahl)
	if err != nil {
		return webx.ServerError(err)
	}

	if result.Created {
		return webx.SeeOther(r, "/dinge/%v/edit", result.Id)
	}

	return webx.SeeOther(r, "/dinge/new")
}

// Zeigt ein spezifisches Ding an
func (a DingeResource) Show(r *http.Request) webx.Response {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return webx.ServerError(err)
	}

	ding, err := a.Repository.GetById(id)
	if err != nil {
		return webx.ServerError(err)
	}

	data := PostDingData{
		Id: id,
		Form: PostDingForm{
			Name:   ding.Name,
			Code:   ding.Code,
			Anzahl: ding.Anzahl,
		},
	}

	template, err := GetTemplate("ding")
	if err != nil {
		return webx.ServerError(err)
	}

	return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
}

// Edit zeigt eine Form zum Bearbeiten eines spezifischen Dings
func (a DingeResource) Edit(r *http.Request) webx.Response {
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
		ValidationErrors validation.ErrorMap
	}{
		Ding: ding,
	}

	template, err := GetTemplate("edit")
	if err != nil {
		return webx.ServerError(err)
	}
	return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
}

func (a DingeResource) Update(r *http.Request) webx.Response {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return webx.NotFound(r)
	}

	form := validation.Form{Request: r}

	var result PostDingForm

	err = form.Scan(
		validation.Field(Name, validation.String(&result.Name), validation.IsNotBlank),
	)

	if err != nil {
		return webx.ServerError(err)
	}

	if !form.IsValid() {
		// a.ServerError(w, r, fmt.Errorf("Validierungsfehler"))
		ding, err := a.Repository.GetById(id)
		if err != nil {
			// Ggf Fehler differenzieren.
			return webx.NotFound(r)
		}

		data := struct {
			Ding             model.Ding
			Validationerrors validation.ErrorMap
		}{
			Ding:             ding,
			Validationerrors: form.ValidationErrors,
		}

		template, err := GetTemplate("edit")
		if err != nil {
			return webx.ServerError(err)
		}

		return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusBadRequest}
	}

	err = a.Repository.NamenAktualisieren(id, result.Name)
	if err != nil {
		return webx.ServerError(err)
	}

	// Im Erfolgsfall zur Datailansicht weiterleiten
	return webx.SeeOther(r, "/dinge/%v", id)
}

func (a DingeResource) Destroy(r *http.Request) webx.Response {

	form := validation.Form{Request: r}

	var anzahl int
	var code string

	err := form.Scan(
		validation.Field("code", validation.String(&code)),
		validation.Field("anzahl", validation.Integer(&anzahl), validation.Min(1)),
	)

	if err != nil {
		return webx.ServerError(err)
	}

	if !form.IsValid() {

		_, err := a.Repository.GetByCode(code)
		if err != nil {
			form.ValidationErrors["code"] = "Unbekannter Produktcode"
			return webx.ServerError(err)
		}

		data := struct {
			Code             string
			Menge            int
			ValidationErrors validation.ErrorMap
		}{
			Code:             code,
			Menge:            anzahl,
			ValidationErrors: form.ValidationErrors,
		}

		template, err := GetTemplate("entnehmen")
		if err != nil {
			return webx.ServerError(err)
		}

		return webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
	}

	id, err := a.Repository.MengeAktualisieren(r.Context(), code, -anzahl)
	if err != nil {
		return webx.ServerError(err)
	}

	return webx.SeeOther(r, "/dinge/%v", id)
}

// Zeigt eine Form an, um Dinge zu entnehmen.
func DestroyForm(r *http.Request) webx.Response {
	template, err := GetTemplate("entnehmen")
	if err != nil {
		return webx.ServerError(err)
	}

	var form = struct {
		Menge int
	}{
		Menge: 1,
	}

	var data = struct {
		Form             struct{ Menge int }
		ValidationErrors validation.ErrorMap
	}{
		Form: form,
	}

	return webx.HtmlResponse{
		Template:   template,
		Data:       data,
		StatusCode: http.StatusOK,
	}
}
