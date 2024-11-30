package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/haschi/dinge/sqlx"
)

type Ding struct {
	DingRef
	Beschreibung string
	Aktualisiert time.Time
}

// DingRef repräsentiert ein Ding in der Übersicht.
type DingRef struct {
	Id     int64
	Name   string
	Code   string
	Anzahl int
}

type Repository struct {
	Clock Clock
	tm    *sqlx.TransactionManager
}

func NewRepository(db *sql.DB, clock Clock) (*Repository, error) {
	tm, err := sqlx.NewTransactionManager(db)
	if err != nil {
		return nil, err
	}

	repository := &Repository{
		Clock: clock,
		tm:    tm,
	}

	return repository, nil
}

var ErrNoRecord = errors.New("no record found")
var ErrInvalidParameter = errors.New("invalid paramater")

func (r Repository) GetById(ctx context.Context, id int64) (Ding, error) {

	if ctx == nil {
		return Ding{}, errors.New("no context provided")
	}

	suchen := `SELECT id, name, code, anzahl, beschreibung, aktualisiert FROM dinge
	WHERE id = ?`

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return Ding{}, err
	}

	defer tx.Rollback()

	// TODO: Das ist hier eine Notlösung. Eigendlich sollte die Logik keine Nullwerte speichern können. Irgendetwas stimmt hier nicht.
	var beschreibung sql.NullString

	var ding Ding
	row := tx.QueryRowContext(ctx, suchen, id)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &beschreibung, &ding.Aktualisiert); err != nil {
		return ding, err
	}

	ding.Beschreibung = beschreibung.String

	return ding, tx.Commit()
}

func (r Repository) GetByCode(ctx context.Context, code string) (Ding, error) {

	if ctx == nil {
		return Ding{}, errors.New("no context provided")
	}

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return Ding{}, err
	}

	defer tx.Rollback()

	suchen := `SELECT id, name, code, anzahl, beschreibung, aktualisiert FROM dinge
	WHERE code = ?`

	var ding Ding
	row := tx.QueryRowContext(ctx, suchen, code)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Beschreibung, &ding.Aktualisiert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Commit()
		}

		return ding, err
	}

	return ding, tx.Commit()
}

func (r Repository) MengeAktualisieren(ctx context.Context, code string, menge int) (*Ding, error) {

	if ctx == nil {
		return nil, errors.New("no context provided")
	}

	tx, err := r.tm.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	statement := `UPDATE dinge
	SET anzahl = anzahl + :anzahl,
	    aktualisiert = :aktualisiert
	WHERE code = :code
	RETURNING id, code, name, anzahl, aktualisiert`

	row := tx.QueryRowContext(ctx, statement,
		sql.Named("code", code),
		sql.Named("anzahl", menge),
		sql.Named("aktualisiert", r.Clock.Now()),
	)

	var ding Ding

	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// TODO: Alle Fehlermeldungen im Protokoll sollen in englisch sein; für Validierungsfehler eigene struct!
			return nil, fmt.Errorf("Unbekannter Produktcode %v: %w)", code, ErrNoRecord)
		}

		return nil, err
	}

	if ding.Anzahl < 0 {
		// Rollback!
		return &ding, fmt.Errorf("Wert ist zu groß: %v: %w", menge, ErrInvalidParameter)
	}

	return &ding, tx.Commit()
}

func (r Repository) Insert(ctx context.Context, code string, anzahl int) (InsertResult, error) {

	var result InsertResult

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return result, err
	}

	defer tx.Rollback()

	statement := `INSERT INTO dinge(name, code, anzahl, beschreibung, aktualisiert)
			VALUES(:name, :code, :anzahl, :beschreibung, :aktualisiert)
			ON CONFLICT (code)
			DO
			UPDATE SET anzahl = anzahl + :anzahl, aktualisiert = :aktualisiert
			WHERE code = :code
			RETURNING id, anzahl`

	timestamp := r.Clock.Now()

	row := tx.QueryRowContext(ctx, statement,
		sql.Named("name", ""),
		sql.Named("code", code),
		sql.Named("anzahl", anzahl),
		sql.Named("beschreibung", ""),
		sql.Named("aktualisiert", timestamp))

	var neueAnzahl int
	if err := row.Scan(&result.Id, &neueAnzahl); err != nil {
		return result, err
	}

	result.Created = neueAnzahl == anzahl

	if result.Created {
		fts := ` INSERT INTO fulltext(rowid, code, name, beschreibung)
		VALUES(:id, :code, '', '')`
		_, err := tx.ExecContext(ctx, fts, sql.Named("id", result.Id), sql.Named("code", code))
		if err != nil {
			return InsertResult{}, err
		}
	}

	return result, tx.Commit()
}

func (r Repository) DingAktualisieren(ctx context.Context, id int64, name string, beschreibung string) error {

	if ctx == nil {
		return errors.New("no context provided")
	}

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// TODO Named Parameter benutzer, keine Fragezeichen.
	statement := `
		UPDATE dinge
		SET name = :name, beschreibung = :beschreibung, aktualisiert = :aktualisiert
		WHERE id = :id;
	`

	timestamp := r.Clock.Now()
	result, err := tx.ExecContext(ctx, statement,
		sql.Named("name", name),
		sql.Named("beschreibung", beschreibung),
		sql.Named("aktualisiert", timestamp),
		sql.Named("id", id))

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected != 1 {
		return ErrNoRecord
	}

	// Update: code kann sich nicht ändern
	updateFulltext := `UPDATE fulltext
	SET name = :name, beschreibung = :beschreibung
	WHERE rowid = :id`

	if _, err := tx.ExecContext(ctx, updateFulltext,
		sql.Named("id", id),
		sql.Named("name", name),
		sql.Named("beschreibung", beschreibung),
	); err != nil {
		return fmt.Errorf("Fehler bei der Abfrage des Volltext-Tabelle: %w", err)
	}

	return tx.Commit()
}

// TODO: Iterator statt Slice zurückgeben.
func (r Repository) GetLatest(ctx context.Context, limit int, query string, sort string) ([]DingRef, error) {

	var statement string

	// Mit Volltextsuche, wenn q nicht leer ist
	if strings.TrimSpace(query) != "" {
		statement = `select id, name, code, anzahl from dinge
	where id IN (SELECT rowid FROM fulltext WHERE fulltext MATCH :query) LIMIT :limit`
	} else {
		// TODO: Named Parameter benutzen.
		statement = `SELECT id, name, code, anzahl FROM dinge
		ORDER BY aktualisiert DESC
		LIMIT :limit`
	}

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return []DingRef{}, err
	}
	defer tx.Rollback()

	if query != "" {
		query = fmt.Sprintf("%v*", query)
	}
	rows, err := tx.QueryContext(ctx, statement, sql.Named("limit", limit), sql.Named("query", query))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var dinge []DingRef = []DingRef{}

	for rows.Next() {
		var ding DingRef
		err = rows.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl)
		if err != nil {
			return dinge, err
		}

		dinge = append(dinge, ding)
	}

	return dinge, tx.Commit()
}

type InsertResult struct {
	Created bool
	Id      int64
}
