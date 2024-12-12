package model

import (
	"database/sql"
	_ "embed"
)

//go:embed create.sql
var CreateScript string

//go:embed fixture.sql
var FixtureScript string

func ExecuteScripts(db *sql.DB, scripts ...string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, script := range scripts {
		_, err := tx.Exec(script)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
