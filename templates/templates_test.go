package templates_test

import (
	"io/fs"
	"slices"
	"testing"

	"github.com/haschi/dinge/templates"
)

func TestTemplates(t *testing.T) {
	sut := templates.TemplatesFileSystem

	entries, err := sut.ReadDir(".")
	if err != nil {
		t.Error(err)
	}

	index := slices.IndexFunc(entries, func(entry fs.DirEntry) bool {
		return entry.Name() == "layout"
	})

	if index == -1 {
		t.Error("layout not found in filesystem")
	}
}
