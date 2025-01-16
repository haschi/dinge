package sqlx

import (
	"database/sql"
)

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

func Execute(scripts ...string) DatabaseInitFunc {
	return func(db *sql.DB) error {
		return ExecuteScripts(db, scripts...)
	}
}

// Funktionen des Typs DatabaseConstructor erzeugen SQL Datenbanken
type DatabaseConstructor func() (*sql.DB, error)

// Funktionen des Type DatabaseInitFunc initialisieren eine Datenbank
type DatabaseInitFunc func(*sql.DB) error
