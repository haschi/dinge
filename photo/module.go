package photo

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/haschi/dinge/validation"
	"github.com/haschi/dinge/webx"

	_ "image/jpeg"
	_ "image/png"
)

type Module struct {
	Repository *Repository
	Templates  fs.FS
}

func (m *Module) Mount(mux *http.ServeMux, prefix string, middleware ...webx.Middleware) {
	mux.Handle(fmt.Sprintf("GET %v/", prefix), webx.CombineFunc(m.Form, middleware...))
	mux.Handle(fmt.Sprintf("POST %v/", prefix), webx.CombineFunc(m.Upload, middleware...))
}

// Form liefert eine Ansicht für die Bearbeitung eines Photos
func (res Module) Form(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	if res.Repository == nil {
		webx.ServerError(w, errors.New("repository not available"))
		return
	}

	url, err := res.Repository.GetUrl(r.Context(), id)

	if err != nil {
		if errors.Is(err, ErrNoRecord) {
			http.NotFound(w, r)
			return
		}
		webx.ServerError(w, err)
		return
	}

	defaultValues := webx.TemplateData[PhotoData]{
		Scripts:          []string{"/static/photo.js"},
		Styles:           []string{"/static/css/photo.css"},
		FormValues:       PhotoData{Id: id, PhotoUrl: url},
		ValidationErrors: nil,
	}

	response := webx.HtmlResponse[PhotoData]{
		TemplateName: "photo",
		Data:         defaultValues,
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, res.Templates); err != nil {
		webx.ServerError(w, err)
	}
}

const Megabyte = 20

// Upload nimmt ein neues Photo entgegen und speichert es in der Datenbank
func (res Module) Upload(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	image, err := imageFromForm(r, "file", 1<<Megabyte)
	if err != nil {

		defaultValues := webx.TemplateData[PhotoData]{
			Scripts:          []string{"/static/photo.js"},
			Styles:           []string{"/static/css/photo.css"},
			FormValues:       PhotoData{Id: id, PhotoUrl: "/static/placeholder.svg"},
			ValidationErrors: validation.ErrorMap{"file": "Fehlerhaft Bilddatei"},
		}

		response := webx.HtmlResponse[PhotoData]{
			TemplateName: "photo",
			Data:         defaultValues,
			StatusCode:   http.StatusUnprocessableEntity,
		}

		if err := response.Render(w, res.Templates); err != nil {
			webx.ServerError(w, err)
		}

		return
	}

	if err := res.Repository.PhotoAktualisieren(r.Context(), id, image); err != nil {
		webx.ServerError(w, err)
		return
	}

	webx.SeeOther("/dinge/%v", id).ServeHTTP(w, r)

}

// Download liefert ein in der Datenbank gespeichertes Photo aus
func (res Module) Download(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	// http.DetectContentType()

	photo, err := res.Repository.GetPhotoById(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNoRecord) {
			http.NotFound(w, r)
			return
		}

		webx.ServerError(w, err)
		return
	}

	reader := bytes.NewReader(photo)
	w.Header().Set("Content-Type", "image/png") // TODO: Den richtigen Mime Type ermitteln
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
}

// TODO: Nach webx verschieben.
func imageFromForm(r *http.Request, field string, limit int64) (image.Image, error) {

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	part, err := reader.NextPart()
	if err != nil {

		return nil, err
	}

	defer part.Close()

	formname := part.FormName()
	if formname != field {
		return nil, fmt.Errorf("unexpected field in multipart form %v", formname)
	}

	// contentType := part.Header.Get("Content-Type")
	// if contentType != mime.

	lr := io.LimitReader(part, limit)

	im, _, err := image.Decode(lr)
	if err != nil {
		return nil, err
	}

	return im, nil

}
