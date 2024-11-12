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

func redirectTo(route string) func(*http.Request) webx.Response {
	return func(r *http.Request) webx.Response {
		return webx.PermanentRedirect(r, route)
	}
}

func handleAbout(r *http.Request) webx.Response {
	template, err := GetTemplate("about")
	if err != nil {
		return webx.ServerError(err)
	}

	return webx.HtmlResponse{Template: template, StatusCode: http.StatusOK}
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
