package sqlx

import "database/sql"

func NewTestDatabase(scripts ...string) (*sql.DB, error) {
	var db *sql.DB
	db, err := sql.Open("sqlite3", inMemoryDataSource)
	if err != nil {
		return db, err
	}

	if err := ExecuteScripts(db, scripts...); err != nil {
		return db, err
	}

	return db, nil
}

var inMemoryDataSource = ConnectionString("test.db", MODE_MEMORY,
	CACHE_Shared,
	JOURNAL_WAL,
	FK_ENABLED)
