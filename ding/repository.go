package ding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/haschi/dinge/sqlx"
)

type Repository struct {
	Clock Clock
	Tm    sqlx.TransactionManager
}

func (r Repository) GetById(ctx context.Context, id int64) (Ding, error) {

	if ctx == nil {
		return Ding{}, errors.New("no context provided")
	}

	suchen := `
	SELECT id, name, code, anzahl, beschreibung, aktualisiert, ('/photos/' || id) AS PhotoUrl
	FROM dinge
	WHERE id = ?
	`

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return Ding{}, err
	}

	defer tx.Rollback()

	// TODO: Das ist hier eine Notlösung. Eigendlich sollte die Logik keine Nullwerte speichern können. Irgendetwas stimmt hier nicht.
	var beschreibung sql.NullString

	var ding Ding
	row := tx.QueryRowContext(suchen, id)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &beschreibung, &ding.Aktualisiert, &ding.PhotoUrl); err != nil {
		return ding, err
	}

	ding.Beschreibung = beschreibung.String

	return ding, tx.Commit()
}

// Todo: Wird nur von Destroy verwendet. Also Spezialisieren!!
func (r Repository) MengeAktualisieren(ctx context.Context, code string, menge int) (*Ding, error) {

	if ctx == nil {
		return nil, errors.New("no context provided")
	}

	tx, err := r.Tm.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	statement := `UPDATE dinge
	SET anzahl = anzahl + :anzahl,
	    aktualisiert = :aktualisiert
	WHERE code = :code
	RETURNING id, code, name, anzahl, aktualisiert`

	row := tx.QueryRowContext(statement,
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

	if err := r.LogEvent(ctx, 3, -menge, ding.Id); err != nil {
		return &ding, err
	}

	return &ding, tx.Commit()
}

func (r Repository) Insert(ctx context.Context, code string, anzahl int) (InsertResult, error) {

	var result InsertResult

	tx, err := r.Tm.BeginTx(ctx)
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

	row := tx.QueryRowContext(statement,
		sql.Named("name", ""),
		sql.Named("code", code),
		sql.Named("anzahl", anzahl),
		sql.Named("beschreibung", ""),
		sql.Named("aktualisiert", timestamp))

	var neueAnzahl int
	if err := row.Scan(&result.Id, &neueAnzahl); err != nil {
		return result, err
	}

	// BUG: Die Annahme, dass der Datensatz neu erzeugt wurde, wenn neueAnzahl == anzahl, ist falsch. Diese Bedingung ist auch wahr, wenn die alte Anzahl == 0 ist. Dadurch kommt es fälschlicherweise zu eine INSERT in die fulltext Tabelle, was einen Constraint Fehler verursacht.
	result.Created = neueAnzahl == anzahl

	if result.Created {
		fts := ` INSERT INTO fulltext(rowid, code, name, beschreibung)
		VALUES(:id, :code, '', '')`
		_, err := tx.ExecContext(fts, sql.Named("id", result.Id), sql.Named("code", code))
		if err != nil {
			return InsertResult{}, err
		}
	}

	var operation int
	if result.Created {
		operation = 1
	} else {
		operation = 2
	}

	if err := r.LogEvent(ctx, operation, anzahl, result.Id); err != nil {
		return InsertResult{}, err
	}

	return result, tx.Commit()
}

func (r Repository) DingAktualisieren(ctx context.Context, id int64, name string, beschreibung string) error {

	if ctx == nil {
		return errors.New("no context provided")
	}

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	statement := `
		UPDATE dinge
		SET name = :name, beschreibung = :beschreibung, aktualisiert = :aktualisiert
		WHERE id = :id;
	`

	timestamp := r.Clock.Now()
	result, err := tx.ExecContext(statement,
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

	if _, err := tx.ExecContext(updateFulltext,
		sql.Named("id", id),
		sql.Named("name", name),
		sql.Named("beschreibung", beschreibung),
	); err != nil {
		return fmt.Errorf("Fehler bei der Abfrage des Volltext-Tabelle: %w", err)
	}

	return tx.Commit()
}

// TODO: Iterator statt Slice zurückgeben.
func (r Repository) Search(ctx context.Context, limit int, query string, sort string) ([]DingRef, error) {

	q := `
		SELECT id, name, code, anzahl, ('/photos/' || id) AS PhotoUrl
			FROM dinge
			WHERE
			  CASE
				  WHEN :query <> ''
					  THEN id IN (SELECT rowid FROM fulltext WHERE fulltext MATCH :query)
						ELSE TRUE
				END
			ORDER BY
				CASE WHEN :sort = 'alpha' THEN name END,
				CASE WHEN :sort = 'omega' THEN name END DESC,
				CASE WHEN :sort = 'oldest' THEN aktualisiert END,
				CASE WHEN :sort = 'latest' THEN aktualisiert END DESC,
				CASE WHEN :sort = '' THEN aktualisiert END DESC
			LIMIT :limit
		`

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return []DingRef{}, err
	}

	defer tx.Rollback()

	rows, err := tx.QueryContext(q,
		sql.Named("limit", limit),
		sql.Named("query", query),
		sql.Named("sort", sort))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var dinge []DingRef = []DingRef{}

	for rows.Next() {
		var ding DingRef
		err = rows.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.PhotoUrl)
		if err != nil {
			return dinge, err
		}

		dinge = append(dinge, ding)
	}

	return dinge, tx.Commit()
}

// ErrDataAccess beschreibt einen Fehler während des Zugriffs auf die Daten des Repositories.
//
// Es handelt sich dabei um einen nicht näher bezeichneten technischen Fehler. Üblicherweise verursachen [context.Canceled] oder [sql.ErrConnDone], [sql.ErrNoRows] oder [sql.ErrTxDone] diesen Fehler.
var ErrDataAccess = errors.New("data access error")

// ErrInsertEvent beschreibt einen Fehler, der beim Protokollieren eines Eregnisses auftritt.
var ErrInsertEvent = errors.New("event cannot be logged")
var ErrNoRecord = errors.New("no record found")
var ErrInvalidParameter = errors.New("invalid paramater")

type InsertResult struct {
	Created bool
	Id      int64
}
