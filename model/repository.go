package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/haschi/dinge/sqlx"
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
	tm    sqlx.TransactionManager
}

func NewRepository(db *sql.DB, clock Clock) Repository {
	return Repository{
		DB:    db,
		Clock: clock,
		tm:    sqlx.TransactionManager{db: db},
	}
}

var ErrNoRecord = errors.New("no record found")

func (r Repository) GetById(ctx context.Context, id int64) (Ding, error) {

	if r.DB == nil {
		return Ding{}, errors.New("no database provided")
	}

	if ctx == nil {
		return Ding{}, errors.New("no context provided")
	}

	suchen := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
	WHERE id = ?`

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return Ding{}, err
	}

	defer tx.Rollback()

	var ding Ding
	row := r.QueryRowContext(ctx, suchen, id)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert); err != nil {
		return ding, err
	}

	return ding, tx.Commit()
}

func (r Repository) GetByCode(ctx context.Context, code string) (Ding, error) {
	if r.DB == nil {
		return Ding{}, errors.New("no database provided")
	}

	if ctx == nil {
		return Ding{}, errors.New("no context provided")
	}

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return Ding{}, err
	}

	defer tx.Rollback()

	suchen := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
	WHERE code = ?`

	var ding Ding
	row := tx.QueryRowContext(ctx, suchen, code)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Commit()
		}

		return ding, err
	}

	return ding, tx.Commit()
}

func (r Repository) MengeAktualisieren(ctx context.Context, code string, menge int) (int64, error) {
	if r.DB == nil {
		return 0, errors.New("no database provided")
	}

	if ctx == nil {
		return 0, errors.New("no context provided")
	}

	tx, err := r.tm.BeginTx(ctx)

	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	ding, err := r.GetByCode(ctx, code)
	if err != nil {
		return 0, err
	}

	if ding.Anzahl+menge < 0 {
		return ding.Id, errors.New("Zuviele")
	}

	if err = r.update(ctx, ding.Id, ding.Anzahl+menge); err != nil {
		return ding.Id, err
	}

	return ding.Id, tx.Commit()
}

func (r Repository) Insert(ctx context.Context, code string, anzahl int) (InsertResult, error) {

	if r.DB == nil {
		return InsertResult{Created: false}, errors.New("no database provided")
	}

	var result InsertResult

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return result, err
	}

	defer tx.Rollback()

	statement := `INSERT INTO dinge(name, code, anzahl, aktualisiert)
			VALUES(:name, :code, :anzahl, :aktualisiert)
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
		sql.Named("aktualisiert", timestamp))

	var neueAnzahl int
	if err := row.Scan(&result.Id, &neueAnzahl); err != nil {
		return result, err
	}

	result.Created = neueAnzahl == anzahl
	return result, tx.Commit()
}

func (r Repository) NamenAktualisieren(ctx context.Context, id int64, name string) error {
	if r.DB == nil {
		return errors.New("no database provided")
	}

	if ctx == nil {
		return errors.New("no context provided")
	}

	tx, err := r.tm.BeginTx(ctx)
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
		return ErrNoRecord
	}

	return tx.Commit()
}

func (r Repository) update(ctx context.Context, id int64, anzahl int) error {

	statement := `UPDATE dinge
	SET anzahl = ?, aktualisiert = ?
	WHERE id = ?`

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

	return tx.Commit()
}

func (r Repository) GetLatest(ctx context.Context, limit int) ([]Ding, error) {
	if r.DB == nil {
		return nil, errors.New("no database provided")
	}

	statement := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
		ORDER BY aktualisiert DESC
		LIMIT ?`

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return []Ding{}, err
	}
	defer tx.Rollback()

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

	return dinge, tx.Commit()
}

type InsertResult struct {
	Created bool
	Id      int64
}

// Clock ist eine Schnittstelle, die das aktuelle Datum und die aktuelle Uhrzeit bereitstellt.
type Clock interface {
	Now() time.Time
}

// RealClock gehört nicht in dieses Paket sondern zum Aufrufer des Paketes (Dependency Inversion)
type RealClock struct{}

func (c RealClock) Now() time.Time {
	return time.Now()
}
