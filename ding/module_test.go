package ding_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/haschi/dinge/ding"
	"github.com/haschi/dinge/sqlx"
	"github.com/haschi/dinge/system"
	"github.com/haschi/dinge/templates"
	"github.com/haschi/dinge/webx"
	"golang.org/x/net/html"
)

func TestModule_GetDinge(t *testing.T) {

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
			args: args{url: "/dinge/"},
			want: want{q: "", s: "", status: http.StatusOK},
		},
		{
			name: "Mit Parameter",
			args: args{url: "/dinge/?q=paprika&s=alpha"},
			want: want{q: "paprika", s: "alpha", status: http.StatusOK},
		},
	}

	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)
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

			q := webx.GetElementById(doc, "input-suche")
			if q == nil {
				t.Error("HTML element not found")
				return
			}

			value := webx.GetAttributeValue(q, "value")
			if value != testcase.want.q {
				t.Errorf("<input id='%v'> value = '%v'; want %v", "q", value, testcase.want.q)
			}

			s := webx.GetElementById(doc, "input-sort")
			selected := webx.GetSelectedOption(s)
			if selected != testcase.want.s {
				t.Errorf("<select id='%v'> selected option value = '%v'; want %v", "s", selected, testcase.want.s)
			}
		})
	}
}

func TestModule_GetDingeNew(t *testing.T) {
	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)

	defer testserver.Close()

	response := testserver.Get("/dinge/new")
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("GET /about want 200; got %v", response.StatusCode)
	}
}

func TestModule_GetDingeId(t *testing.T) {

	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)

	defer testserver.Close()

	tests := []struct {
		name string
		arg  int64
		want int
	}{
		{
			name: "Show known thing",
			arg:  1,
			want: http.StatusOK,
		},
		{
			name: "Show unknown thing",
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

func TestModule_GetDingeIdEdit(t *testing.T) {
	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)
	defer testserver.Close()

	tests := []struct {
		name string
		arg  int64
		want int
	}{
		{
			name: "Edit an existing thing",
			arg:  1,
			want: http.StatusOK,
		},
		{
			name: "Edit unknown thing",
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

func TestModule_PostDinge(t *testing.T) {

	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)

	defer testserver.Close()

	type fixture struct {
		name           string
		data           url.Values
		wantStatusCode int
		wantLocation   string
	}

	tests := []fixture{
		{
			name: "Add new things",
			data: url.Values{
				ding.Code:   []string{"42"},
				ding.Anzahl: []string{"7"},
			},
			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/4/edit",
		},
		{
			name: "Add known things",
			data: url.Values{
				ding.Code:   []string{"111"},
				ding.Anzahl: []string{"7"},
			},
			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/new",
		},
		{
			name: "Add known things with invalid data.",
			data: url.Values{
				ding.Code:   []string{"111"},
				ding.Anzahl: []string{"invalid number"},
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			resp := testserver.Post("/dinge/", test.data)
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

func TestModule_PostDingeId(t *testing.T) {
	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)
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
			name: "Rename a known thing.",
			path: "/dinge/1",
			data: url.Values{
				ding.Name: []string{"Salat"},
			},

			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/1",
		},
		{
			name:           "Rename thing with malformed identity identifier.",
			path:           "/dinge/malformed",
			data:           url.Values{},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Rename unknown thing",
			path: "/dinge/42",
			data: url.Values{
				ding.Name: []string{"Kenn ich nicht"},
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Renaming with invalid names for a thing",
			path: "/dinge/1",
			data: url.Values{
				ding.Name: []string{""},
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

func TestModule_PostDingeDelete(t *testing.T) {
	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)
	defer testserver.Close()

	type fixture struct {
		name           string
		data           url.Values
		wantStatusCode int
		wantLocation   string
	}

	tests := []fixture{
		{
			name: "Remove a familiar thing",
			data: url.Values{
				ding.Code:   []string{"111"},
				ding.Anzahl: []string{"1"},
			},

			wantStatusCode: http.StatusSeeOther,
			wantLocation:   "/dinge/1",
		},
		{
			name: "Remove thing with malformed identity identifier",
			data: url.Values{
				ding.Code:   []string{"malfomed"},
				ding.Anzahl: []string{"1"},
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Remove several known things",
			data: url.Values{
				ding.Code:   []string{"111"},
				ding.Anzahl: []string{"3"},
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
func TestModule_GetDingeDelete(t *testing.T) {
	config := newTestConfig()
	testserver := webx.NewTestserver(t, "/dinge", config)
	defer testserver.Close()

	path := "/dinge/delete"
	response := testserver.Get(path)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("GET %v = %v; want %v", path, response.StatusCode, http.StatusOK)
	}

}

func newDingTestModule(initFncs ...sqlx.DatabaseInitFunc) webx.ModuleConstructor {
	return func(db *sql.DB) (webx.Module, error) {
		for _, fn := range initFncs {
			if err := fn(db); err != nil {
				return nil, err
			}
		}

		tm, err := sqlx.NewSqlTransactionManager(db)
		if err != nil {
			return nil, err
		}

		repository := &ding.Repository{Clock: system.RealClock{}, Tm: tm}

		module := &ding.Module{
			Repository: repository,
			Templates:  templates.TemplatesFileSystem,
		}

		return module, nil
	}
}

func newTestConfig() webx.TestserverConfig {
	scripts := sqlx.Execute(ding.CreateScript, ding.FixtureScript)
	config := webx.TestserverConfig{
		Database:   webx.InMemoryDatabase(),
		Module:     newDingTestModule(scripts),
		Middleware: []webx.Middleware{},
	}

	return config
}
