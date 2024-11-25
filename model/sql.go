package model

import (
	"database/sql"
	_ "embed"

	"github.com/haschi/dinge/testx"
)

//go:embed create.sql
var CreateScript string

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(CreateScript)
	return err
}

//go:embed fixture.sql
var FixtureScript string

func SetupFixture(db *sql.DB) error {
	_, err := db.Exec(FixtureScript)
	return err
}

func RunScript(script string) testx.SetupFunc {
	return func(d *sql.DB) error {
		_, err := d.Exec(script)
		return err
	}
}

func ExecuteScripts(db *sql.DB, scripts ...string) error {
	for _, script := range scripts {
		_, err := db.Exec(script)
		if err != nil {
			return err
		}
	}
	return nil
}
