package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/haschi/dinge/model"
)

func TestResource_GetAbout(t *testing.T) {

	testserver := newTestServer(t, newDingeResource)

	defer testserver.Close()

	response := testserver.Get("/about")
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("GET /about want 200; got %v", response.StatusCode)
	}
}

func TestResource_GetDinge(t *testing.T) {
	testserver := newTestServer(t, newDingeResource)

	defer testserver.Close()

	response := testserver.Get("/dinge")
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("GET /about want 200; got %v", response.StatusCode)
	}
}

func TestResource_GetDingeNew(t *testing.T) {
	testserver := newTestServer(t, newDingeResource)

	defer testserver.Close()

	response := testserver.Get("/dinge/new")
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("GET /about want 200; got %v", response.StatusCode)
	}
}

func TestResource_GetDingeId(t *testing.T) {

	testserver := newTestServer(t, newDingeResource)

	defer testserver.Close()

	tests := []struct {
		name string
		arg  int64
		want int
	}{
		{
			name: "GET existing id",
			arg:  1,
			want: http.StatusOK,
		},
		{
			name: "GET unknown id",
			arg:  42,
			want: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("/dinge/%v", test.arg)
			response := testserver.Get(path)
			defer response.Body.Close()

			if response.StatusCode != test.want {
				t.Errorf("GET %v = %v; want %v", path, response.StatusCode, test.want)
			}
		})
	}
}

func TestResource_GetDingeIdEdit(t *testing.T) {
	testserver := newTestServer(t, newDingeResource)

	defer testserver.Close()

	tests := []struct {
		name string
		arg  int64
		want int
	}{
		{
			name: "GET existing id",
			arg:  1,
			want: http.StatusOK,
		},
		{
			name: "GET unknown id",
			arg:  42,
			want: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("/dinge/%v/edit", test.arg)
			response := testserver.Get(path)
			defer response.Body.Close()

			if response.StatusCode != test.want {
				t.Errorf("GET %v = %v; want %v", path, response.StatusCode, test.want)
			}
		})
	}
}

func TestResource_GetDingeDelete(t *testing.T) {
	testserver := newTestServer(t, newDingeResource)

	defer testserver.Close()

	path := "/dinge/delete"
	response := testserver.Get(path)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("GET %v = %v; want %v", path, response.StatusCode, http.StatusOK)
	}

}

const dataSource = "file::memory:?cache=shared"

type testserver struct {
	t      *testing.T
	db     *sql.DB
	server *httptest.Server
}

func newTestServer(t *testing.T, fn func(db *sql.DB) DingeResource) *testserver {
	t.Helper()

	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		t.Fatal(err)
	}

	scripts := []string{model.CreateScript, model.FixtureScript}
	if err := model.ExecuteScripts(db, scripts...); err != nil {
		t.Fatal(err)
	}

	var loglevel = new(slog.LevelVar)
	loghandler := slog.NewJSONHandler(os.Stdin, &slog.HandlerOptions{Level: loglevel})
	logger := slog.New(loghandler)

	resource := fn(db)
	testserver := &testserver{
		db:     db,
		server: httptest.NewTLSServer(routes(logger, resource)),
	}

	client := testserver.server.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return testserver
}

func (t *testserver) Close() {
	if t.server != nil {
		t.server.Close()
	}

	if t.db != nil {
		if err := t.db.Close(); err != nil {
			t.t.Fatal(err)
		}
	}
}

func (t *testserver) Get(path string) *http.Response {
	client := t.server.Client()
	resp, err := client.Get(t.server.URL + path)
	if err != nil {
		t.t.Fatal(err)
	}
	return resp
}

func newDingeResource(db *sql.DB) DingeResource {
	return DingeResource{
		Repository: model.Repository{
			DB:    db,
			Clock: model.RealClock{},
		},
	}
}
