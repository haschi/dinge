package main

import (
	"net/http"

	"github.com/haschi/dinge/webx"
)

const (
	Name         = "name"
	Anzahl       = "anzahl"
	Code         = "code"
	Beschreibung = "beschreibung"
)

func handleAbout(w http.ResponseWriter, r *http.Request) {
	response := webx.HtmlResponse[any]{
		TemplateName: "about",
		Data:         webx.TemplateData[any]{},
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, TemplatesFileSystem); err != nil {
		webx.ServerError(w, err)
	}
}
