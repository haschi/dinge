package photo_test

import (
	"database/sql"
	"testing"

	"github.com/haschi/dinge/photo"
	"github.com/haschi/dinge/sqlx"
	"github.com/haschi/dinge/system"
	"github.com/haschi/dinge/templates"
	"github.com/haschi/dinge/webx"
)

func TestResource_Form(t *testing.T) {

	type args struct {
		url string
	}

	type want struct {
		statusCode int
	}

	type testcase struct {
		name string
		args args
		want want
	}

	testcases := []testcase{
		{
			name: "new photo",
			args: args{url: "/dinge/3/photo"},
			want: want{statusCode: 200},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {

			scripts := sqlx.Execute(photo.CreateScript)
			config := webx.TestserverConfig{
				Database:   webx.InMemoryDatabase(),
				Module:     NewPhotoTestModule(scripts),
				Middleware: []webx.Middleware{},
			}

			testserver := webx.NewTestserver(t, "/dinge/{id}", config)
			defer testserver.Close()

			response := testserver.Get(testcase.args.url)
			defer response.Body.Close()

			if response.StatusCode != testcase.want.statusCode {
				t.Errorf("GET %v want status %v, got %v", testcase.args.url, testcase.want.statusCode, response.StatusCode)
			}
		})
	}
}

func NewPhotoTestModule(initFncs ...sqlx.DatabaseInitFunc) webx.ModuleConstructor {
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

		repository := &photo.Repository{Clock: system.RealClock{}, Tm: tm}

		module := &photo.Module{
			Templates:  templates.TemplatesFileSystem,
			Repository: repository,
		}

		return module, nil
	}
}
