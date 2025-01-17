package ding

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/haschi/dinge/validation"
	"github.com/haschi/dinge/webx"
)

type Module struct {
	Repository *Repository
	Templates  fs.FS
	Photos     webx.Module
}

func (m *Module) Mount(mux *http.ServeMux, prefix string, middleware ...webx.Middleware) {
	mux.Handle(fmt.Sprintf("GET %v/", prefix), webx.CombineFunc(m.Index, middleware...))
	mux.Handle(fmt.Sprintf("GET %v/new", prefix), webx.CombineFunc(m.NewForm, middleware...))
	mux.Handle(fmt.Sprintf("POST %v/", prefix), webx.CombineFunc(m.Create, middleware...))
	mux.Handle(fmt.Sprintf("GET %v/{id}", prefix), webx.CombineFunc(m.Show, middleware...))
	mux.Handle(fmt.Sprintf("GET %v/{id}/edit", prefix), webx.CombineFunc(m.Edit, middleware...))
	mux.Handle(fmt.Sprintf("POST %v/{id}", prefix), webx.CombineFunc(m.Update, middleware...))
	mux.Handle(fmt.Sprintf("GET %v/delete", prefix), webx.CombineFunc(m.DestroyForm, middleware...))
	mux.Handle(fmt.Sprintf("POST %v/delete", prefix), webx.CombineFunc(m.Destroy, middleware...))
}

// Zeigt eine Liste aller Dinge
func (m Module) Index(w http.ResponseWriter, r *http.Request) {

	content := webx.TemplateData[IndexFormData]{}

	form := validation.NewForm(r)
	defer form.Close()

	sortOptions := validation.StringOptions("alpha", "omega", "latest", "oldest")

	form.Scan(
		validation.String("q", &content.FormValues.Q, validation.MaxLength(100)),
		validation.String("s", &content.FormValues.S, sortOptions),
	)

	dinge, err := m.Repository.Search(r.Context(), 12, content.FormValues.Q, content.FormValues.S)
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	content.FormValues.Result = dinge

	// TODO: Validierungsfehler in index.tmpl anzeigen
	response := webx.HtmlResponse[IndexFormData]{
		TemplateName: "index",
		Data:         content,
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, m.Templates); err != nil {
		webx.ServerError(w, err)
	}
}

// Liefert eine HTML Form zum Einlagern eines neuen Dings.
func (m Module) NewForm(w http.ResponseWriter, r *http.Request) {

	history, err := m.Repository.GetAllEvents(r.Context(), 12)
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	data := NewScannerFormData("", 1, history)
	response := webx.HtmlResponse[ScannerFormData]{
		TemplateName: "new",
		Data:         data,
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, m.Templates); err != nil {
		webx.ServerError(w, err)
	}
}

// Lagert Dinge ein.
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
func (m Module) Create(w http.ResponseWriter, r *http.Request) {

	form := validation.NewForm(r)
	defer form.Close()

	data := NewScannerFormData("", 0, nil)
	data.ValidationErrors = form.ValidationErrors

	err := form.Scan(
		validation.String(Code, &data.FormValues.Code, validation.IsNotBlank),
		validation.Integer(Anzahl, &data.FormValues.Anzahl, validation.Min(1)),
	)

	if err != nil {
		webx.ServerError(w, err)
		return
	}

	if !form.IsValid() {
		// "fehler in den übermittelten Daten"

		response := webx.HtmlResponse[ScannerFormData]{
			TemplateName: "new",
			Data:         data,
			StatusCode:   http.StatusUnprocessableEntity,
		}

		if err := response.Render(w, m.Templates); err != nil {
			webx.ServerError(w, err)
			return
		}

		return
	}

	// TODO: data.FormValues übergeben.
	result, err := m.Repository.Insert(r.Context(), data.FormValues.Code, data.FormValues.Anzahl)
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

// Zeigt ein spezifisches Ding an
func (m Module) Show(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	ding, err := m.Repository.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		webx.ServerError(w, err)
		return
	}

	history, err := m.Repository.ProductHistory(r.Context(), id, 10)
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	data := webx.TemplateData[ShowResponseData]{
		FormValues: ShowResponseData{
			Ding:    ding,
			History: history,
		},
	}

	response := webx.HtmlResponse[ShowResponseData]{
		TemplateName: "ding",
		Data:         data,
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, m.Templates); err != nil {
		webx.ServerError(w, err)
	}
}

// Edit zeigt eine Form zum Bearbeiten eines spezifischen Dings
func (m Module) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	ding, err := m.Repository.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		webx.ServerError(w, err)
		return
	}

	data := webx.TemplateData[Ding]{
		FormValues: ding,
	}

	response := webx.HtmlResponse[Ding]{
		TemplateName: "edit",
		Data:         data,
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, m.Templates); err != nil {
		webx.ServerError(w, err)
	}
}

// Bearbeitet ein Ding
func (m Module) Update(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	form := validation.NewForm(r)
	defer form.Close()

	var name string
	var beschreibung string

	err = form.Scan(
		validation.String(Name, &name, validation.IsNotBlank),
		validation.String(Beschreibung, &beschreibung),
	)

	if err != nil {
		webx.ServerError(w, err)
		return
	}

	if !form.IsValid() {
		// m.ServerError(w, r, fmt.Errorf("Validierungsfehler"))
		ding, err := m.Repository.GetById(r.Context(), id)
		if err != nil {
			// TODO: Fehler differenzieren.
			http.NotFound(w, r)
			return
		}

		data := webx.TemplateData[Ding]{
			FormValues:       ding,
			ValidationErrors: form.ValidationErrors,
		}

		response := webx.HtmlResponse[Ding]{
			TemplateName: "edit",
			Data:         data,
			StatusCode:   http.StatusUnprocessableEntity,
		}

		if err := response.Render(w, m.Templates); err != nil {
			webx.ServerError(w, err)
			return
		}

		return
	}

	err = m.Repository.DingAktualisieren(r.Context(), id, name, beschreibung)
	if err != nil {
		if errors.Is(err, ErrNoRecord) {
			http.NotFound(w, r)
			return
		}
		webx.ServerError(w, err)
		return
	}

	// Im Erfolgsfall zur Datailansicht weiterleiten
	// TODO: Der Präfix darf nicht hart codiert werden, damit das Modul nicht an einen konkreten Pfad gebunden ist und verschoben werden kann.
	webx.SeeOther("/dinge/%v", id).ServeHTTP(w, r)
}

func (m Module) Destroy(w http.ResponseWriter, r *http.Request) {

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

	var ding *Ding
	if form.IsValid() {
		ding, err = m.Repository.MengeAktualisieren(r.Context(), code, -anzahl)
		if err != nil {
			switch {
			case errors.Is(err, ErrNoRecord):
				form.ValidationErrors[Code] = "Unbekannter Produktcode"
			case errors.Is(err, ErrInvalidParameter):
				form.ValidationErrors[Anzahl] = "Anzahl zu groß"
			default:
				webx.ServerError(w, err)
				return
			}
		}
	}

	if !form.IsValid() {

		history, err := m.Repository.GetAllEvents(r.Context(), 12)
		if err != nil {
			webx.ServerError(w, err)
			return
		}
		data := NewDestroyFormData(code, anzahl, history)
		data.ValidationErrors = form.ValidationErrors
		response := webx.HtmlResponse[ScannerFormData]{

			TemplateName: "entnehmen",
			Data:         data,
			StatusCode:   http.StatusUnprocessableEntity,
		}

		if err := response.Render(w, m.Templates); err != nil {
			webx.ServerError(w, err)
			return
		}

		return
	}

	webx.SeeOther("/dinge/%v", ding.Id).ServeHTTP(w, r)
}

// Zeigt eine Form an, um Dinge zu entnehmen.
func (m Module) DestroyForm(w http.ResponseWriter, r *http.Request) {
	history, err := m.Repository.GetAllEvents(r.Context(), 12)
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	data := NewDestroyFormData("", 1, history)

	response := webx.HtmlResponse[ScannerFormData]{
		TemplateName: "entnehmen",
		Data:         data,
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, m.Templates); err != nil {
		webx.ServerError(w, err)
	}
}
