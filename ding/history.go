package ding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

// ProductHistory liefert eine Liste der Ereignisse zu einem Ding
//
// TODO: Es muss differenziert werden zwischen der Historie eines Dings und der Historie in der Übersicht. Die Historie zu einem Ding benötigt keine Referenz zum Ding selbst, das dieses ja bekannt ist. Hingegen benötigt die Übersicht aber eben jene Referenz, damit eine Navigation möglich ist.
func (r Repository) ProductHistory(ctx context.Context, dingId int64, limit int) ([]Event, error) {

	q := `
	SELECT id, code, name, operation, count, created
	FROM history
	INNER JOIN dinge ON history.dinge_id = dinge.id
	WHERE dinge.id = :id
	ORDER BY created DESC
	LIMIT :limit
	`

	return r.GetEvents(ctx, q, sql.Named("id", dingId), sql.Named("limit", limit))
}

func (r Repository) GetEvents(ctx context.Context, query string, args ...any) ([]Event, error) {
	history := []Event{}

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return history, err
	}

	defer tx.Rollback()

	rows, err := tx.QueryContext(query, args...)
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
			&event.Anzahl,
			&event.Created); err != nil {
			return history, err
		}

		history = append(history, event)
	}

	return history, nil
}

func (r Repository) GetAllEvents(ctx context.Context, limit int) ([]Event, error) {
	q := `
		SELECT id, code, name, operation, count, created
		FROM history
		INNER JOIN dinge ON history.dinge_id = dinge.id
		ORDER BY created DESC
		LIMIT :limit
		`

	return r.GetEvents(ctx, q, sql.Named("limit", limit))
}

type Event struct {
	Operation int
	Anzahl    int
	Created   time.Time
	DingRef
}

func (e Event) Equal(other Event) bool {
	return e.Operation == other.Operation &&
		e.Anzahl == other.Anzahl &&
		e.Created.Equal(other.Created) &&
		e.DingRef.Equal(other.DingRef)
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

type Comperable[T any] interface {
	Equal(other T) bool
}

func SliceEqual[T Comperable[T]](left []T, right []T) bool {
	if len(left) != len(right) {
		return false
	}

	for index, v := range left {
		if !v.Equal(right[index]) {
			return false
		}
	}

	return true
}
