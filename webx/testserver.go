package webx

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/haschi/dinge/sqlx"
)

func NewTestserver(t *testing.T, prefix string, options TestserverConfig) *Testserver {

	t.Helper()

	db, err := options.Database()
	if err != nil {
		t.Fatal(err)
	}

	mod, err := options.Module(db)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()

	mod.Mount(mux, prefix, options.Middleware...)

	testserver := &Testserver{
		db:     db,
		t:      t,
		module: mod,
		server: httptest.NewTLSServer(mux),
	}

	client := testserver.server.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return testserver
}

func (s *Testserver) Close() {
	if s.server != nil {
		s.server.Close()
	}

	if s.db != nil {
		s.db.Close()
	}
}

func (t *Testserver) Get(path string) *http.Response {
	t.t.Helper()

	client := t.server.Client()
	resp, err := client.Get(t.server.URL + path)
	if err != nil {
		t.t.Fatal(err)
	}
	return resp
}

func (t *Testserver) Post(path string, data url.Values) *http.Response {

	t.t.Helper()

	url := t.server.URL + path
	resp, err := t.server.Client().PostForm(url, data)
	if err != nil {
		t.t.Fatal(err)
	}

	return resp
}

type Testserver struct {
	t      *testing.T
	server *httptest.Server
	module Module
	db     *sql.DB
}

type TestserverConfig struct {
	Database   sqlx.DatabaseConstructor
	Module     ModuleConstructor
	Middleware []Middleware
}

func NoOpInit(*sql.DB) error { return nil }

// InMemoryDatabase erzeugt eine Funktion, die eine SQLite3 Datanbank im Speicher erzeugt
//
// Nach dem Erstellen der Datanbank werden die inits Funktionen mit der erzeugten Datenbank als Parameter, um die Datenbank zu initialisieren; zum Beispiel ein Schema anzulegen.
func InMemoryDatabase(inits ...sqlx.DatabaseInitFunc) sqlx.DatabaseConstructor {
	return func() (*sql.DB, error) {
		db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
		if err != nil {
			return db, err
		}

		for _, i := range inits {
			if err := i(db); err != nil {
				return db, err
			}
		}

		return db, nil
	}
}

type ModuleConstructor func(db *sql.DB) (Module, error)
