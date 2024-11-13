package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Ding struct {
	Id           int64
	Name         string
	Code         string
	Anzahl       int
	Aktualisiert time.Time
}

type Repository struct {
	*sql.DB
	Clock Clock
}

func (r Repository) GetById(id int64) (Ding, error) {
	if r.DB == nil {
		return Ding{}, errors.New("no database provided")
	}

	suchen := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
	WHERE id = ?`

	var ding Ding
	row := r.QueryRow(suchen, id)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert); err != nil {
		return ding, err
	}

	return ding, nil
}

func (r Repository) GetByCode(code string) (Ding, error) {
	if r.DB == nil {
		return Ding{}, errors.New("no database provided")
	}

	suchen := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
	WHERE code = ?`

	var ding Ding
	row := r.QueryRow(suchen, code)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert); err != nil {
		return ding, err
	}

	return ding, nil
}

func (r Repository) MengeAktualisieren(ctx context.Context, code string, menge int) (int64, error) {
	if r.DB == nil {
		return 0, errors.New("no database provided")
	}

	var id int64

	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return id, err
	}

	defer tx.Rollback()

	suchen := `SELECT id, anzahl FROM dinge
	WHERE code = ?`
	var alteAnzahl int

	row := tx.QueryRowContext(ctx, suchen, code)
	if err := row.Scan(&id, &alteAnzahl); err != nil {
		return id, err
	}

	if alteAnzahl+menge < 0 {
		return id, errors.New("Zuviele")
	}

	if err = r.update(ctx, tx, id, alteAnzahl+menge); err != nil {
		return id, err
	}

	return id, tx.Commit()
}

func (r Repository) Insert(ctx context.Context, code string, anzahl int) (InsertResult, error) {

	if r.DB == nil {
		return InsertResult{Created: false}, errors.New("no database provided")
	}

	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return InsertResult{Created: false}, err
	}

	defer tx.Rollback()

	suchen := `SELECT id, anzahl FROM dinge
	WHERE code = ?`

	row := tx.QueryRowContext(ctx, suchen, code)

	var alteAnzahl int
	var id int64
	if err := row.Scan(&id, &alteAnzahl); err != nil {

		// Prüfen, ob ErrNoRows. Dann das da:
		statement := `INSERT INTO dinge(name, code, anzahl, aktualisiert)
		VALUES(?, ?, ?, ?)`

		timestamp := r.Clock.Now()
		result, err := tx.ExecContext(ctx, statement, "", code, anzahl, timestamp)
		if err != nil {
			return InsertResult{}, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return InsertResult{}, err
		}

		return InsertResult{Created: true, Id: id}, tx.Commit()

		// sonst Fehler: Form nochmal anzeigen mit Fehlermeldung.
	}

	// Ding ist schon vorhanden und muss aktualisiert werden.
	// TODO: Prüfen, dass alteAnzahl+anzahl >= 0 !
	if err := r.update(ctx, tx, id, alteAnzahl+anzahl); err != nil {
		return InsertResult{}, err
	}

	return InsertResult{Created: false, Id: id}, tx.Commit()
}

func (r Repository) NamenAktualisieren(ctx context.Context, id int64, name string) error {
	if r.DB == nil {
		return errors.New("no database provided")
	}

	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	statement := `UPDATE dinge
	SET name = ?, aktualisiert = ?
	WHERE id = ?`

	timestamp := r.Clock.Now()
	result, err := tx.ExecContext(ctx, statement, name, timestamp, id)
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

	return tx.Commit()
}

func (r Repository) update(ctx context.Context, tx *sql.Tx, id int64, anzahl int) error {

	statement := `UPDATE dinge
	SET anzahl = ?, aktualisiert = ?
	WHERE id = ?`

	result, err := tx.ExecContext(ctx, statement, anzahl, r.Clock.Now(), id)
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

func (r Repository) GetLatest(limit int) ([]Ding, error) {
	if r.DB == nil {
		return nil, errors.New("no database provided")
	}

	statement := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
		ORDER BY aktualisiert DESC
		LIMIT ?`

	rows, err := r.Query(statement, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var dinge []Ding = []Ding{}

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

type InsertResult struct {
	Created bool
	Id      int64
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (c RealClock) Now() time.Time {
	return time.Now()
}
