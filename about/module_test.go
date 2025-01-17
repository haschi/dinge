package about_test

import (
	"database/sql"
	"net/http"
	"slices"
	"testing"

	"github.com/haschi/dinge/about"
	"github.com/haschi/dinge/templates"
	"github.com/haschi/dinge/webx"
	"golang.org/x/net/html"

	_ "github.com/mattn/go-sqlite3"
)

func TestAboutModule_GetLicense(t *testing.T) {

	config := webx.TestserverConfig{
		Database:   webx.InMemoryDatabase(),
		Module:     NewAboutTestModule(),
		Middleware: []webx.Middleware{},
	}

	testserver := webx.NewTestserver(t, "/about", config)
	defer testserver.Close()

	url := "/about/license"

	response := testserver.Get(url)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("GET %v want 200; got %v", url, response.StatusCode)
	}

	doc, err := html.Parse(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	nodes := webx.GetElement(doc, "h2")
	if !slices.ContainsFunc(nodes, func(e *html.Node) bool {
		return e.FirstChild != nil && e.FirstChild.Data == "License"
	}) {
		t.Error("Element <h2>License</h2> ist nicht enthalten")
	}
}

func NewAboutTestModule() webx.ModuleConstructor {
	return func(db *sql.DB) (webx.Module, error) {
		return &about.Module{
			Templates: templates.TemplatesFileSystem,
		}, nil
	}
}
