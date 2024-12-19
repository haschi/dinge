package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

func (r Repository) LogEvent(ctx context.Context, operation int, count int, dingId int64) error {
	statement := `INSERT INTO history(operation, count, created, dinge_id)
	VALUES(:operation, :count, :created, :dinge_id)`

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return dataAccessError(err)
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(statement,
		sql.Named("operation", operation),
		sql.Named("count", count),
		sql.Named("created", r.Clock.Now()),
		sql.Named("dinge_id", dingId),
	)

	if err != nil {
		var sqlError sqlite3.Error
		if errors.As(err, &sqlError) {
			if sqlError.Code == sqlite3.ErrConstraint {
				return ErrInvalidParameter
			}
		}

		return dataAccessError(err)
	}

	return dataAccessError(tx.Commit())
}

func dataAccessError(err error) error {
	if err == nil {
		return nil
	}

	// TODO: Ursprünglichen Fehler protokollieren
	return ErrDataAccess
}

// ErrDataAccess beschreibt einen Fehler während des Zugriffs auf die Daten des Repositories.
//
// Es handelt sich dabei um einen nicht näher bezeichneten technischen Fehler. Üblicherweise verursachen [context.Canceled] oder [sql.ErrConnDone], [sql.ErrNoRows] oder [sql.ErrTxDone] diesen Fehler.
var ErrDataAccess = errors.New("data access error")

// ErrInsertEvent beschreibt einen Fehler, der beim Protokollieren eines Eregnisses auftritt.
var ErrInsertEvent = errors.New("event cannot be logged")

func (r Repository) GetHistory(ctx context.Context, limit int) ([]Event, error) {
	q := `
		SELECT id, code, name, operation, count
		FROM history
		INNER JOIN dinge ON history.dinge_id = dinge.id
		ORDER BY created DESC
		LIMIT :limit
		`

	history := []Event{}

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return history, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(q, sql.Named("limit", limit))
	if err != nil {
		return history, err
	}

	for rows.Next() {
		event := Event{}
		if err := rows.Scan(
			&event.DingRef.Id,
			&event.DingRef.Code,
			&event.DingRef.Name,
			&event.Operation,
			&event.Anzahl); err != nil {
			return history, err
		}
		history = append(history, event)
	}

	return history, nil
}

type Event struct {
	Operation int
	Anzahl    int
	DingRef
}

func (e Event) String() string {
	switch e.Operation {
	case 1:
		return fmt.Sprintf("Neu erfasst: %v Stück", e.Anzahl)
	case 2:
		return fmt.Sprintf("Eingelagert: %v Stück", e.Anzahl)
	case 3:
		return fmt.Sprintf("Entnommen: %v Stück", e.Anzahl)
	default:
		return "Unbekannte Operation"
	}
}
