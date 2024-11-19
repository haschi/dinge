package main

import (
	"fmt"
	"html/template"
	"net/http"

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

func handleAbout(w http.ResponseWriter, r *http.Request) {
	template, err := GetTemplate("about")
	if err != nil {
		webx.ServerError(w, err)
		return
	}

	response := webx.HtmlResponse{Template: template, StatusCode: http.StatusOK}
	if err := response.Render(w); err != nil {
		webx.ServerError(w, err)
	}
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
