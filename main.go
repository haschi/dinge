/*
	Dinge startet den Dinge Dienst

	Usage:

	  dinge [flags]

Die flags sind:

	  --address
		    Legt fest, an welche Adresse der Server gebunden wird.
				Die Addresse besteht aus einem Interface und einem Port, die durch einen Doppelpunkt getrennt werden.
				Mit "127.0.0.1:8080" wird der Server zum Beispiel an Lokalhost Port 8080 gebunden.
				Wenn als Interface 0.0.0.0 gewählt wird, bindet der Server an alle Interfaces. Die Voreinstellung ist "0.0.0.0:8443", falls dieser Parameter nicht angegeben wird.

		--datasorce
		    Der Name der Datenquelle, die zum Speichern der Daten des Services benutzt wird.
				Im einfachsten Fall handelt es sich um den Pfad zu einer SQLite Datenbankdatei.
				Wenn die Datenbankdatei nicht existiert, wird sie angelegt.
				Es können zusätzliche Parameter für die Datenquelle angegegebn werden.
				Sie dazu auch https://github.com/mattn/go-sqlite3?tab=readme-ov-file#connection-string

		--version -v
		    Gibt die Version aus.
				Der Server wird nicht gestartet.
*/
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/haschi/dinge/model"
	_ "github.com/mattn/go-sqlite3"
)

var Git_Revision string
var Version string = "development"

func main() {
	ctx := context.Background()

	if err := run(ctx, os.Stdout, os.Args, os.LookupEnv); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer, _ []string, environment func(string) (string, bool)) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	httpAddress := environmentOrDefault(environment, "HTTP_ADDRESS", "0.0.0.0:8443")
	flag.StringVar(&httpAddress, "address", httpAddress, "HTTP network address")

	datasource := environmentOrDefault(environment, "DATASOURCE", "dinge.db")
	flag.StringVar(&datasource, "datasource", datasource, "SQLite data source name")

	var version bool
	flag.BoolVar(&version, "version", false, "print version information")
	flag.BoolVar(&version, "v", false, "print version information (shorthand)")
	flag.Parse()

	if version {
		// TODO: Richtig machen!
		fmt.Println("Dinge", Git_Revision, Version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(stdout, nil))
	logger.Info("starting server", slog.String("address", httpAddress))

	db, err := sql.Open("sqlite3", datasource)
	if err != nil {
		logger.Error("Can not open database",
			slog.String("source", err.Error()),
			slog.String("datasource", datasource))
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("closing database", slog.String("source", err.Error()))
		}
		logger.Info("database closed")
	}()

	if strings.Contains(datasource, ":memory:") {
		logger.Error("datasource :memory: not supported. use ?mode=memory instead.", slog.String("datasource", datasource))
		os.Exit(1)
	}

	stmt := "SELECT sqlite_version()"
	var db_version string
	db.QueryRow(stmt).Scan(&db_version)

	logger.Info("using database",
		slog.String("version", db_version),
		slog.String("datasource", datasource))

	if err != nil {
		return err
	}

	application := DingeApplication{
		Repository: model.Repository{DB: db},
		Logger:     logger,
	}

	server := &http.Server{
		Addr:     httpAddress,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelInfo),
		Handler:  routes(application),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		logger.Info("starting http server", slog.String("address", httpAddress))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error listening and serving", slog.String("source", err.Error()))
		}
		logger.Info("Stopped serving new connections")
	}()

	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 20*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("error shutting down http server",
				slog.String("source", err.Error()))
		}

		logger.Info("graceful shutdown complete")
	}()

	wg.Wait()
	logger.Info("stopped")

	return nil
}

func environmentOrDefault(environment func(string) (string, bool), key string, defaultValue string) string {
	if environmentValue, ok := environment(key); ok {
		return environmentValue
	}

	return defaultValue
}

type DingeApplication struct {
	Logger     *slog.Logger
	Repository model.Repository
}

func (a DingeApplication) Error(w http.ResponseWriter, r *http.Request, err error) {
	a.Logger.Error(err.Error(),
		slog.String("method", r.Method),
		slog.String("uri", r.URL.RequestURI()))

	http.Error(w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func routes(dinge DingeApplication) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", dinge.HandleGet)

	mux.HandleFunc("POST /{$}", dinge.HandlePost)

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	mux.HandleFunc("GET /{id}", dinge.HandleGetDing)

	mux.HandleFunc("POST /{id}", dinge.HandlePostDing)
	return mux
}

type Form struct {
	Code   string
	Anzahl int
}

type Data struct {
	LetzteEinträge []model.Ding
	Form           Form
	FieldErrors    map[string]string
}
