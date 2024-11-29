package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/system"
	"golang.org/x/net/html"
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

	type args struct {
		url string
	}

	type want struct {
		q      string
		s      string
		status int
	}

	type testcase struct {
		name string
		args args
		want want
	}

	testcases := []testcase{
		{
			name: "Ohne Parameter",
			args: args{url: "/dinge"},
			want: want{q: "", s: "", status: http.StatusOK},
		},
		{
			name: "Mit Parameter",
			args: args{url: "/dinge?q=paprika&s=alpha"},
			want: want{q: "paprika", s: "alpha", status: http.StatusOK},
		},
	}

	testserver := newTestServer(t, newDingeResource)
	defer testserver.Close()

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {

			response := testserver.Get(testcase.args.url)
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Errorf("GET /about want 200; got %v", response.StatusCode)
				return
			}

			doc, err := html.Parse(response.Body)
			if err != nil {
				t.Fatal(err)
			}

			q := getById(doc, "input-suche")
			if q == nil {
				t.Error("HTML element not found")
				return
			}

			value := getAttributeValue(q, "value")
			if value != testcase.want.q {
				t.Errorf("<input id='%v'> value = '%v'; want %v", "q", value, testcase.want.q)
			}

			s := getById(doc, "input-sort")
			selected := getSelectedOption(s)
			if selected != testcase.want.s {
				t.Errorf("<select id='%v'> selected option value = '%v'; want %v", "s", selected, testcase.want.s)
			}
		})
	}
}

func getSelectedOption(n *html.Node) string {
	for descendant := range n.Descendants() {
		if descendant.Type == html.ElementNode && descendant.Data == "option" {
			if hasAttribute(descendant, "selected") {
				return getAttributeValue(descendant, "value")
			}
		}
	}
	return ""
}

func hasAttribute(n *html.Node, key string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return true
		}
	}

	return false
}

func getAttributeValue(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}

func getById(node *html.Node, id string) *html.Node {

	if getAttributeValue(node, "id") == id {
		return node
	}

	next := node.NextSibling
	for next != nil {
		if result := getById(next, id); result != nil {
			return result
		}
		next = next.NextSibling
	}

	for child := range node.Descendants() {
		return getById(child, id)
	}

	return nil
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
			wantStatusCode: http.StatusUnprocessableEntity,
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
	logger *slog.Logger
}

func newTestServer(t *testing.T, fn func(*slog.Logger, *sql.DB) (http.Handler, error)) *testserver {
	t.Helper()

	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		t.Fatal(err)
	}

	scripts := []string{model.CreateScript, model.FixtureScript}
	if err := model.ExecuteScripts(db, scripts...); err != nil {
		t.Fatal(err)
	}

	var buffer bytes.Buffer
	var loglevel = new(slog.LevelVar)
	loghandler := slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: loglevel})
	logger := slog.New(loghandler)

	resource, err := fn(logger, db)
	if err != nil {
		t.Fatal(err)
	}

	testserver := &testserver{
		db:     db,
		server: httptest.NewTLSServer(resource),
		logger: logger,
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

func newDingeResource(logger *slog.Logger, db *sql.DB) (http.Handler, error) {
	repository, err := model.NewRepository(db, system.RealClock{})
	if err != nil {
		return nil, err
	}
	resource := &DingeResource{
		Repository: repository,
	}

	return routes(logger, *resource), nil
}
