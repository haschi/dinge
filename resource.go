package main

import (
	"fmt"
	"html/template"
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

func (a DingeResource) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	a.Logger.Error(err.Error(),
		slog.String("method", r.Method),
		slog.String("uri", r.URL.RequestURI()))

	http.Error(w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

// Was hier fehlt ist die Middleware Kette, so dass Fehler ggf. von
// dem umschließenden Handler geloggt werden können.

// Liefert eine HTML Form zum Erzeugen eines neuen Dings.
func (a DingeResource) NewForm(w http.ResponseWriter, r *http.Request) {

	var index, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/new/*.tmpl")

	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	form := struct {
		Code   string
		Anzahl int
	}{
		Code:   "",
		Anzahl: 1,
	}

	data := struct {
		Form struct {
			Code   string
			Anzahl int
		}
		ValidationErrors validation.ErrorMap
	}{
		Form: form,
	}

	if err := render(w, http.StatusOK, index, data); err != nil {
		a.ServerError(w, r, err)
		return
	}
}

// Zeigt eine Liste aller Dinge
func (a DingeResource) Index(w http.ResponseWriter, r *http.Request) {

	// Kommt nach DingeApplication
	var index, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/index/*.tmpl")

	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	dinge, err := a.Repository.GetLatest()
	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	data := Data{
		LetzteEinträge: dinge,
		Form:           Form{Anzahl: 1},
	}

	if err := render(w, http.StatusOK, index, data); err != nil {
		a.ServerError(w, r, err)
		return
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

	var index, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/index/*.tmpl")

	if err != nil {
		a.ServerError(w, r, err)
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
		a.ServerError(w, r, err)
		return
	}

	if !form.IsValid() {
		// a.Error(w, r, fmt.Errorf("fehler in den übermittelten Daten"))

		dinge, err := a.Repository.GetLatest()
		if err != nil {
			a.ServerError(w, r, err)
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

	result, err := a.Repository.Insert(r.Context(), code, anzahl)
	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	a.Logger.Info("Neues Ding erzeugt", slog.Int64("id", result.Id), slog.Bool("created", result.Created))

	if result.Created {
		http.Redirect(w, r, fmt.Sprintf("/dinge/%v/edit", result.Id), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dinge/new", http.StatusSeeOther)
}

// Zeigt ein spezifisches Ding an
func (a DingeResource) Show(w http.ResponseWriter, r *http.Request) {

	var page, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/ding/*.tmpl")

	if err != nil {
		a.ServerError(w, r, err)
		return
	}

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

	data := PostDingData{
		Id: id,
		Form: PostDingForm{
			Name:   ding.Name,
			Code:   ding.Code,
			Anzahl: ding.Anzahl,
		},
	}

	if err := render(w, http.StatusOK, page, data); err != nil {
		a.ServerError(w, r, err)
		return
	}
}

// Edit zeigt eine Form zum Bearbeiten eines spezifischen Dings
func (a DingeResource) Edit(w http.ResponseWriter, r *http.Request) {
	var page, err = template.ParseFS(
		TemplatesFileSystem,
		"templates/layout/*.tmpl",
		"templates/pages/edit/*.tmpl")

	if err != nil {
		a.ServerError(w, r, err)
		return
	}

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
		ValidationErrors validation.ErrorMap
	}{
		Ding: ding,
	}

	if err := render(w, http.StatusOK, page, data); err != nil {
		a.ServerError(w, r, err)
		return
	}
}

func (a DingeResource) Update(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	form := validation.Form{Request: r}

	var result PostDingForm

	err = form.Scan(
		validation.Field(Name, validation.String(&result.Name), validation.IsNotBlank),
		// validation.Field(Code, validation.String(&result.Code), validation.IsNotBlank),
		// validation.Field(Anzahl, validation.Integer(&result.Anzahl), validation.Min(1)),
	)

	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	if !form.IsValid() {
		a.ServerError(w, r, fmt.Errorf("Validierungsfehler"))
		ding, err := a.Repository.GetById(id)
		if err != nil {
			// Ggf Fehler differenzieren.
			http.NotFound(w, r)
			return
		}

		page, err := template.ParseFS(
			TemplatesFileSystem,
			"templates/layout/*.tmpl",
			"templates/pages/edit/*.tmpl")

		if err != nil {
			a.ServerError(w, r, err)
			return
		}

		data := struct {
			Ding             model.Ding
			Validationerrors validation.ErrorMap
		}{
			Ding:             ding,
			Validationerrors: form.ValidationErrors,
		}

		// Bei einem Validierungsfehler die Form erneut anzeigen, mit den
		// eingegebenen Werten vorbelegen und die Fehlermeldungen ausgeben.
		if err := render(w, http.StatusOK, page, data); err != nil {
			a.ServerError(w, r, err)
			return
		}

		return
	}

	err = a.Repository.NamenAktualisieren(id, result.Name)
	if err != nil {
		a.ServerError(w, r, err)
	}

	// Im Erfolgsfall zur Datailansicht weiterleiten
	http.Redirect(w, r, fmt.Sprintf("/dinge/%v", id), http.StatusSeeOther)
}

func (a DingeResource) Destroy(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		a.ServerError(w, r, err)
		return
	}

	form := validation.Form{Request: r}

	var menge int

	err = form.Scan(
		validation.Field("menge", validation.Integer(&menge), validation.Min(1)),
	)

	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	if !form.IsValid() {
		page, err := template.ParseFS(
			TemplatesFileSystem,
			"templates/layout/*.tmpl",
			"templates/pages/entnehmen/menge/*.tmpl")

		if err != nil {
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
			Ding:             ding,
			Menge:            1,
			ValidationErrors: form.ValidationErrors,
		}

		if err := render(w, http.StatusOK, page, data); err != nil {
			a.ServerError(w, r, err)
			return
		}
	}

	err = a.Repository.MengeAktualisieren(r.Context(), id, -menge)
	if err != nil {
		a.ServerError(w, r, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%v", id), http.StatusSeeOther)
}
