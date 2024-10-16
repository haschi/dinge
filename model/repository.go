package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Ding struct {
	Id           int
	Name         string
	Code         string
	Anzahl       int
	Aktualisiert time.Time
}

type Repository struct {
	*sql.DB
}

func (r Repository) GetById(id int64) (Ding, error) {
	suchen := `SELECT id, name, code, anzahl FROM dinge
	WHERE id = ?`

	var ding Ding
	row := r.QueryRow(suchen, id)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl); err != nil {
		return ding, err
	}

	return ding, nil
}

func (r Repository) Insert(ctx context.Context, code string, anzahl int) (int64, error) {

	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	suchen := `SELECT id, anzahl FROM dinge
	WHERE code = ?`

	var alteAnzahl int
	var id int64
	row := r.QueryRow(suchen, code)

	if err := row.Scan(&id, &alteAnzahl); err != nil {

		statement := `INSERT INTO dinge(name, code, anzahl, aktualisiert)
		VALUES(?, ?, ?, datetime('now', 'utc'))`

		result, err := r.Exec(statement, "", code, anzahl)
		if err != nil {
			return 0, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}

		tx.Commit()
		return id, nil
	}

	if err := r.Update(id, alteAnzahl+anzahl); err != nil {
		return 0, err
	}

	tx.Commit()
	return id, nil
}

func (r Repository) NamenAktualisieren(id int64, name string) error {
	statement := `UPDATE dinge
	SET name = ?, aktualisiert = datetime('now', 'utc')
	WHERE id = ?`

	result, err := r.Exec(statement, name, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected != 1 {
		return fmt.Errorf("expected 1 row affected, got %v", affected)
	}

	return nil
}

func (r Repository) Update(id int64, anzahl int) error {

	statement := `UPDATE dinge
	SET anzahl = ?, aktualisiert = datetime('now', 'utc')
	WHERE id = ?`

	result, err := r.Exec(statement, anzahl, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected != 1 {
		return fmt.Errorf("expected 1 row affected, got %v", affected)
	}

	return nil
}

func (r Repository) GetLatest() ([]Ding, error) {

	statement := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
		ORDER BY aktualisiert DESC
		LIMIT 10`

	rows, err := r.Query(statement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var dinge []Ding
	for rows.Next() {
		var ding Ding
		err = rows.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert)
		if err != nil {
			return dinge, err
		}

		dinge = append(dinge, ding)
	}

	return dinge, nil
}
