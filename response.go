package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Renderer func(w http.ResponseWriter) error

type ResourceHandler func(a DingeResource, r *http.Request) Renderer

func SeeOther(r *http.Request, url string) Renderer {
	return func(w http.ResponseWriter) error {
		http.Redirect(w, r, url, http.StatusSeeOther)
		return nil
	}
}

func ServerError(err error) Renderer {
	return func(w http.ResponseWriter) error {
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return err
	}
}

func NotFound(r *http.Request) Renderer {
	return func(w http.ResponseWriter) error {
		http.NotFound(w, r)
		return nil
	}
}

func HtmlResponse(view string, data any, statusCode int) Renderer {
	return func(w http.ResponseWriter) error {
		page, err := template.ParseFS(
			TemplatesFileSystem,
			"templates/layout/*.tmpl",
			fmt.Sprintf("templates/pages/%v/*.tmpl", view))

		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)

			return err
		}

		if err = render(w, statusCode, page, data); err != nil {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)
			return err
		}

		return nil
	}
}
