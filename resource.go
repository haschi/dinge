package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/validation"
)

type DingeResource struct {
	Logger     *slog.Logger
	Repository model.Repository
}

func ResponseWrapper(logger *slog.Logger, handler func(r *http.Request) Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := handler(r)
		if err := response(w); err != nil {
			logger.Error(err.Error(),
				slog.String("method", r.Method),
				slog.String("uri", r.URL.RequestURI()))
		}
	}
}

// Was hier fehlt ist die Middleware Kette, so dass Fehler ggf. von
// dem umschließenden Handler geloggt werden können.

// Liefert eine HTML Form zum Erzeugen eines neuen Dings.
func (a DingeResource) NewForm(r *http.Request) Renderer {

	data := FormData[CreateData]{
		Form: CreateData{Anzahl: 1},
	}

	return HtmlResponse("new", data, http.StatusOK)
}

// Zeigt eine Liste aller Dinge
func (a DingeResource) Index(r *http.Request) Renderer {

	dinge, err := a.Repository.GetLatest()
	if err != nil {
		return ServerError(err)
	}

	data := Data{
		LetzteEinträge: dinge,
		Form:           Form{Anzahl: 1},
	}

	return HtmlResponse("index", data, http.StatusOK)
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
func (a DingeResource) Create(r *http.Request) Renderer {

	form := validation.Form{Request: r}

	var content CreateData

	err := form.Scan(
		validation.Field(Code, validation.String(&content.Code), validation.IsNotBlank),
		validation.Field(Anzahl, validation.Integer(&content.Anzahl), validation.Min(1)),
	)

	if err != nil {
		return ServerError(err)
	}

	if !form.IsValid() {
		// "fehler in den übermittelten Daten"

		data := FormData[CreateData]{
			Form:             content,
			ValidationErrors: form.ValidationErrors,
		}

		return HtmlResponse("new", data, http.StatusUnprocessableEntity)
	}

	result, err := a.Repository.Insert(r.Context(), content.Code, content.Anzahl)
	if err != nil {
		return ServerError(err)
	}

	if result.Created {
		return SeeOther(r, fmt.Sprintf("/dinge/%v/edit", result.Id))
	}

	return SeeOther(r, "/dinge/new")
}

// Zeigt ein spezifisches Ding an
func (a DingeResource) Show(r *http.Request) Renderer {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return ServerError(err)
	}

	ding, err := a.Repository.GetById(id)
	if err != nil {
		return ServerError(err)
	}

	data := PostDingData{
		Id: id,
		Form: PostDingForm{
			Name:   ding.Name,
			Code:   ding.Code,
			Anzahl: ding.Anzahl,
		},
	}

	return HtmlResponse("ding", data, http.StatusOK)
}

// Edit zeigt eine Form zum Bearbeiten eines spezifischen Dings
func (a DingeResource) Edit(r *http.Request) Renderer {
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
		ValidationErrors validation.ErrorMap
	}{
		Ding: ding,
	}

	return HtmlResponse("edit", data, http.StatusOK)
}

func (a DingeResource) Update(r *http.Request) Renderer {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return NotFound(r)
	}

	form := validation.Form{Request: r}

	var result PostDingForm

	err = form.Scan(
		validation.Field(Name, validation.String(&result.Name), validation.IsNotBlank),
	)

	if err != nil {
		return ServerError(err)
	}

	if !form.IsValid() {
		// a.ServerError(w, r, fmt.Errorf("Validierungsfehler"))
		ding, err := a.Repository.GetById(id)
		if err != nil {
			// Ggf Fehler differenzieren.
			return NotFound(r)
		}

		data := struct {
			Ding             model.Ding
			Validationerrors validation.ErrorMap
		}{
			Ding:             ding,
			Validationerrors: form.ValidationErrors,
		}

		return HtmlResponse("edit", data, http.StatusBadRequest)
	}

	err = a.Repository.NamenAktualisieren(id, result.Name)
	if err != nil {
		return ServerError(err)
	}

	// Im Erfolgsfall zur Datailansicht weiterleiten
	return SeeOther(r, fmt.Sprintf("/dinge/%v", id))
}

func (a DingeResource) Destroy(r *http.Request) Renderer {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return ServerError(err)
	}

	form := validation.Form{Request: r}

	var menge int

	err = form.Scan(
		validation.Field("menge", validation.Integer(&menge), validation.Min(1)),
	)

	if err != nil {
		return ServerError(err)
	}

	if !form.IsValid() {

		ding, err := a.Repository.GetById(id)
		if err != nil {
			return ServerError(err)
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

		return HtmlResponse("entnehmen/menge", data, http.StatusOK)
	}

	err = a.Repository.MengeAktualisieren(r.Context(), id, -menge)
	if err != nil {
		return ServerError(err)
	}

	return SeeOther(r, fmt.Sprintf("/%v", id))
}
