package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/validation"
	"github.com/haschi/dinge/webx"
)

type DingeResource struct {
	Repository *model.Repository
}

// Zeigt eine Liste aller Dinge
func (a DingeResource) Index(w http.ResponseWriter, r *http.Request) {

	dinge, err := a.Repository.GetLatest(r.Context(), 12)
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	content := IndexContent{
		Dinge: dinge,
	}

	template, err := GetTemplate("index")
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	response := webx.HtmlResponse{Template: template, Data: content, StatusCode: http.StatusOK}
	if err := response.Render(w); err != nil {
		webx.ServerError(w, err)
	}
}

type IndexFormData struct {
}

type IndexContent struct {
	Dinge []model.Ding
	Form  validation.FormData[IndexFormData]
}

// Liefert eine HTML Form zum Erzeugen eines neuen Dings.
func (a DingeResource) NewForm(w http.ResponseWriter, r *http.Request) {

	data := validation.FormData[CreateData]{
		Form: CreateData{Anzahl: 1},
	}

	template, err := GetTemplate("new")
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	response := webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
	if err := response.Render(w); err != nil {
		webx.ServerError(w, err)
	}
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
func (a DingeResource) Create(w http.ResponseWriter, r *http.Request) {

	form := validation.NewForm(r)
	defer form.Close()

	var content CreateData

	err := form.Scan(
		validation.String(Code, &content.Code, validation.IsNotBlank),
		validation.Integer(Anzahl, &content.Anzahl, validation.Min(1)),
	)

	if err != nil {
		webx.ServerError(w, err)
		return
	}

	if !form.IsValid() {
		// "fehler in den übermittelten Daten"

		data := validation.FormData[CreateData]{
			Form:             content,
			ValidationErrors: form.ValidationErrors,
		}

		template, err := GetTemplate("new")
		if err != nil {
			webx.ServerError(w, err)
			return
		}

		response := webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusUnprocessableEntity}
		if err := response.Render(w); err != nil {
			webx.ServerError(w, err)
			return
		}
	}

	result, err := a.Repository.Insert(r.Context(), content.Code, content.Anzahl)
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	if result.Created {
		webx.SeeOther("/dinge/%v/edit", result.Id).ServeHTTP(w, r)
		return
	}

	webx.SeeOther("/dinge/new").ServeHTTP(w, r)
}

type CreateData struct {
	Code   string
	Anzahl int
}

// Zeigt ein spezifisches Ding an
func (a DingeResource) Show(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	ding, err := a.Repository.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		webx.ServerError(w, err)
		return
	}

	data := validation.FormData[model.Ding]{
		Form: ding,
	}

	template, err := GetTemplate("ding")
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	response := webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
	if err := response.Render(w); err != nil {
		webx.ServerError(w, err)
	}
}

// Edit zeigt eine Form zum Bearbeiten eines spezifischen Dings
func (a DingeResource) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	ding, err := a.Repository.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		webx.ServerError(w, err)
		return
	}

	data := validation.FormData[model.Ding]{
		Form: ding,
	}

	template, err := GetTemplate("edit")
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	response := webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusOK}
	if err := response.Render(w); err != nil {
		webx.ServerError(w, err)
	}
}

func (a DingeResource) Update(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	form := validation.NewForm(r)
	defer form.Close()

	var result PostDingForm

	err = form.Scan(
		validation.String(Name, &result.Name, validation.IsNotBlank),
	)

	if err != nil {
		webx.ServerError(w, err)
		return
	}

	if !form.IsValid() {
		// a.ServerError(w, r, fmt.Errorf("Validierungsfehler"))
		ding, err := a.Repository.GetById(r.Context(), id)
		if err != nil {
			// TODO: Fehler differenzieren.
			http.NotFound(w, r)
			return
		}

		data := validation.FormData[model.Ding]{
			Form:             ding,
			ValidationErrors: form.ValidationErrors,
		}

		template, err := GetTemplate("edit")
		if err != nil {
			webx.ServerError(w, err)
			return
		}

		response := webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusUnprocessableEntity}
		if err := response.Render(w); err != nil {
			webx.ServerError(w, err)
			return
		}

		return
	}

	err = a.Repository.NamenAktualisieren(r.Context(), id, result.Name)
	if err != nil {
		if errors.Is(err, model.ErrNoRecord) {
			http.NotFound(w, r)
			return
		}
		webx.ServerError(w, err)
		return
	}

	// Im Erfolgsfall zur Datailansicht weiterleiten
	webx.SeeOther("/dinge/%v", id).ServeHTTP(w, r)
}

func (a DingeResource) Destroy(w http.ResponseWriter, r *http.Request) {

	form := validation.NewForm(r)

	var anzahl int
	var code string

	err := form.Scan(
		validation.String(Code, &code, validation.IsNotBlank),
		validation.Integer(Anzahl, &anzahl, validation.Min(1)))

	if err != nil {
		webx.ServerError(w, err)
		return
	}

	var ding *model.Ding
	if form.IsValid() {
		ding, err = a.Repository.MengeAktualisieren(r.Context(), code, -anzahl)
		if err != nil {
			switch {
			case errors.Is(err, model.ErrNoRecord):
				form.ValidationErrors[Code] = "Unbekannter Produktcode"
			case errors.Is(err, model.ErrInvalidParameter):
				form.ValidationErrors[Anzahl] = "Anzahl zu groß"
			default:
				webx.ServerError(w, err)
				return
			}
		}
	}

	if !form.IsValid() {

		data := validation.FormData[DestroyData]{
			Form: DestroyData{
				Code:  code,
				Menge: anzahl,
			},
			ValidationErrors: form.ValidationErrors,
		}

		template, err := GetTemplate("entnehmen")
		if err != nil {
			webx.ServerError(w, err)
			return
		}

		response := webx.HtmlResponse{Template: template, Data: data, StatusCode: http.StatusUnprocessableEntity}
		if err := response.Render(w); err != nil {
			webx.ServerError(w, err)
			return
		}

		return
	}

	webx.SeeOther("/dinge/%v", ding.Id).ServeHTTP(w, r)
}

// Zeigt eine Form an, um Dinge zu entnehmen.
func DestroyForm(w http.ResponseWriter, r *http.Request) {
	template, err := GetTemplate("entnehmen")
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	data := validation.FormData[DestroyData]{
		Form: DestroyData{
			Code:  "",
			Menge: 1,
		},
	}

	response := webx.HtmlResponse{
		Template:   template,
		Data:       data,
		StatusCode: http.StatusOK,
	}

	if err := response.Render(w); err != nil {
		webx.ServerError(w, err)
	}
}

type DestroyData struct {
	Code  string
	Menge int
}
