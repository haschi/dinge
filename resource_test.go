package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/system"
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

func TestResource_PostDinge(t *testing.T) {

	testserver := newTestServer(t, newDingeResource)
	defer testserver.Close()

	type fixture struct {
		name           string
		data           url.Values
		wantStatusCode int
		wantLocation   string
	}

	tests := []fixture{
		{
			name: "Neues Ding",
			data: url.Values{
				Code:   []string{"42"},
				Anzahl: []string{"7"},
			},
			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/4/edit",
		},
		{
			name: "Ding aktualisieren",
			data: url.Values{
				Code:   []string{"111"},
				Anzahl: []string{"7"},
			},
			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/new",
		},
		{
			name: "Ungültige Daten",
			data: url.Values{
				Code:   []string{"111"},
				Anzahl: []string{"invalid number"},
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			resp := testserver.Post("/dinge", test.data)
			if resp.StatusCode != test.wantStatusCode {
				t.Errorf("POST /dinge = %v; want %v", resp.StatusCode, test.wantStatusCode)
			}

			location := resp.Header.Get("Location")
			if location != test.wantLocation {
				t.Errorf("Header Location = %v, want %v", location, test.wantLocation)
			}
		})
	}
}

func TestResource_PostDingeId(t *testing.T) {
	testserver := newTestServer(t, newDingeResource)
	defer testserver.Close()

	type fixture struct {
		name           string
		path           string
		data           url.Values
		wantStatusCode int
		wantLocation   string
	}

	tests := []fixture{
		{
			name: "Aktualisierunge Ding",
			path: "/dinge/1",
			data: url.Values{
				Name: []string{"Salat"},
			},

			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/1",
		},
		{
			name:           "Missgebildeter Identitätsbezeichner",
			path:           "/dinge/malformed",
			data:           url.Values{},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Nicht vorhandenes Ding aktualisieren",
			path: "/dinge/42",
			data: url.Values{
				Name: []string{"Kenn ich nicht"},
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Ungültige Daten",
			path: "/dinge/1",
			data: url.Values{
				Name: []string{""},
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := testserver.Post(test.path, test.data)

			gotStatusCode := resp.StatusCode
			if gotStatusCode != test.wantStatusCode {
				t.Errorf("POST %v = %v, want %v", test.path, gotStatusCode, test.wantStatusCode)
			}

			location := resp.Header.Get("Location")
			if location != test.wantLocation {
				t.Errorf("POST %v = %v, want %v", test.path, location, test.wantLocation)
			}
		})
	}
}

func TestResource_PostDingeDelete(t *testing.T) {
	testserver := newTestServer(t, newDingeResource)
	defer testserver.Close()

	type fixture struct {
		name           string
		data           url.Values
		wantStatusCode int
		wantLocation   string
	}

	tests := []fixture{
		{
			name: "Entnehme ein Paprika",
			data: url.Values{
				Code:   []string{"111"},
				Anzahl: []string{"1"},
			},

			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/1",
		},
		{
			name: "Ungültiger Code",
			data: url.Values{
				Code:   []string{"malfomed"},
				Anzahl: []string{"1"},
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Entnehme drei Paprika",
			data: url.Values{
				Code:   []string{"111"},
				Anzahl: []string{"3"},
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := "/dinge/delete"
			resp := testserver.Post(path, test.data)

			gotStatusCode := resp.StatusCode
			if gotStatusCode != test.wantStatusCode {
				t.Errorf("POST %v = %v, want %v", path, gotStatusCode, test.wantStatusCode)
			}

			location := resp.Header.Get("Location")
			if location != test.wantLocation {
				t.Errorf("POST %v = %v, want %v", path, location, test.wantLocation)
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

func newTestServer(t *testing.T, fn func(db *sql.DB) (*DingeResource, error)) *testserver {
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

	resource, err := fn(db)
	if err != nil {
		t.Fatal(err)
	}

	testserver := &testserver{
		db:     db,
		server: httptest.NewTLSServer(routes(logger, *resource)),
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

func (t *testserver) Post(path string, data url.Values) *http.Response {
	url := t.server.URL + path
	resp, err := t.server.Client().PostForm(url, data)
	if err != nil {
		t.t.Fatal(err)
	}
	defer resp.Body.Close()

	return resp
}

func newDingeResource(db *sql.DB) (*DingeResource, error) {
	repository, err := model.NewRepository(db, system.RealClock{})
	if err != nil {
		return nil, err
	}
	resource := &DingeResource{
		Repository: repository,
	}

	return resource, nil
}
