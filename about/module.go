package about

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/haschi/dinge/webx"
)

func (res Module) GetLicense(w http.ResponseWriter, r *http.Request) {
	response := webx.HtmlResponse[any]{
		TemplateName: "license",
		Data:         webx.TemplateData[any]{},
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, res.Templates); err != nil {
		webx.ServerError(w, err)
	}
}

func (res Module) GetUsage(w http.ResponseWriter, r *http.Request) {
	response := webx.HtmlResponse[any]{
		TemplateName: "usage",
		Data:         webx.TemplateData[any]{},
		StatusCode:   http.StatusOK,
	}

	if err := response.Render(w, res.Templates); err != nil {
		webx.ServerError(w, err)
	}
}

type Module struct {
	Templates embed.FS
}

func (m *Module) Mount(mux *http.ServeMux, prefix string, middleware ...webx.Middleware) {
	mux.Handle(fmt.Sprintf("GET %v/license", prefix), webx.CombineFunc(m.GetLicense, middleware...))
	mux.Handle(fmt.Sprintf("GET %v/usage", prefix), webx.CombineFunc(m.GetUsage, middleware...))
}

func (m *Module) Close() error {
	return nil
}
