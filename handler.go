package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/haschi/dinge/model"
)

// Was hier fehlt ist die Middleware Kette, so dass Fehler ggf. von
// dem umschließenden Handler geloggt werden können.
func handleGet(_ *slog.Logger, repository model.Repository) applicationHandler {

	var index, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/index/*.tmpl")

	return func(w http.ResponseWriter, r *http.Request) *HandlerError {
		if err != nil {
			return &HandlerError{
				Message:    "template parsing error",
				StatusCode: http.StatusInternalServerError,
				Source:     err,
			}
		}

		var buffer bytes.Buffer

		dinge, err := repository.GetLatest()
		if err != nil {
			return &HandlerError{
				Message:    "repository error",
				Source:     err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		data := Data{
			LetzteEinträge: dinge,
			Form:           Form{Anzahl: 1},
		}

		err = index.Execute(&buffer, data)
		if err != nil {
			return &HandlerError{
				Message:    "template execution error",
				Source:     err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		w.WriteHeader(http.StatusOK)
		buffer.WriteTo(w)

		return nil
	}
}

func handlePost(logger *slog.Logger, repository model.Repository) applicationHandler {

	var _, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/index/*.tmpl")

	return func(w http.ResponseWriter, r *http.Request) *HandlerError {
		if err != nil {
			return &HandlerError{
				Message:    "template parsing error",
				Source:     err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		err = r.ParseForm()
		if err != nil {
			return &HandlerError{
				Message:    "form parsing error",
				Source:     err,
				StatusCode: http.StatusBadRequest,
			}
		}

		var code = r.PostForm.Get("code")
		var anzahl_str = r.PostForm.Get("anzahl")
		anzahl, err := strconv.Atoi(anzahl_str)
		if err != nil {
			return &HandlerError{
				Message:    "can not convert to integer: 'anzahl'",
				Source:     err,
				StatusCode: http.StatusUnprocessableEntity,
			}
		}

		logger.Info("got post form",
			slog.String("code", code),
			slog.Int("anzahl", anzahl))

		id, err := repository.Insert(r.Context(), code, anzahl)
		if err != nil {
			return &HandlerError{
				Message:    "can not insert record",
				Source:     err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		logger.Info("insert database record", slog.Int64("id", id))

		// dinge, err := repository.GetLatest()
		// if err != nil {
		// 	return &HandlerError{
		// 		Message:    "error reading from repository",
		// 		Source:     err,
		// 		StatusCode: http.StatusInternalServerError,
		// 	}
		// }

		// Im Fehlerfall die Form mit den alten Werten füllen
		// und Fehlermeldungen hinzufügen.
		// var buffer bytes.Buffer

		// err = index.Execute(&buffer, Data{
		// 	LetzteEinträge: dinge,
		// 	Form: Form{
		// 		Anzahl: anzahl,
		// 		Code:   code,
		// 	},
		// })

		// if err != nil {
		// 	return &HandlerError{
		// 		Message:    "template execution error",
		// 		Source:     err,
		// 		StatusCode: http.StatusInternalServerError,
		// 	}
		// }

		// w.WriteHeader(http.StatusOK)
		// buffer.WriteTo(w)

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return nil
	}
}

func handleGetDing(repository model.Repository) applicationHandler {

	var page, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/ding/*.tmpl")

	return func(w http.ResponseWriter, r *http.Request) *HandlerError {
		if err != nil {
			return &HandlerError{
				Message:    "error parsing template",
				Source:     err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			http.NotFound(w, r)
			return nil
		}

		ding, err := repository.GetById(id)
		if err != nil {
			return &HandlerError{
				Message:    "not found",
				Source:     err,
				StatusCode: http.StatusNotFound,
			}
		}

		var buffer bytes.Buffer
		if err := page.Execute(&buffer, ding); err != nil {
			return &HandlerError{
				Message:    "error executing template",
				Source:     err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		w.WriteHeader(http.StatusOK)
		buffer.WriteTo(w)
		return nil
	}
}

func handlePostDing(repository model.Repository) applicationHandler {
	var _, err = template.ParseFS(
		Templates,
		"templates/layout/*.tmpl",
		"templates/pages/ding/*.tmpl")

	return func(w http.ResponseWriter, r *http.Request) *HandlerError {
		if err != nil {
			return &HandlerError{
				Message:    "error parsing template",
				Source:     err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			http.NotFound(w, r)
			return nil
		}

		err = r.ParseForm()
		if err != nil {
			return &HandlerError{
				Message:    "form parsing error",
				Source:     err,
				StatusCode: http.StatusBadRequest,
			}
		}

		var name = r.PostForm.Get("name")

		err = repository.NamenAktualisieren(id, name)
		if err != nil {
			return &HandlerError{
				Message:    "not found",
				Source:     err,
				StatusCode: http.StatusNotFound,
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/%v", id), http.StatusSeeOther)
		return nil
	}
}
