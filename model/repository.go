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
	//*sql.DB
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

	suchen := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
	WHERE id = ?`

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return Ding{}, err
	}

	defer tx.Rollback()

	var ding Ding
	row := tx.QueryRowContext(ctx, suchen, id)
	if err := row.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert); err != nil {
		return ding, err
	}

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

func (r Repository) GetLatest(ctx context.Context, limit int) ([]Ding, error) {

	statement := `SELECT id, name, code, anzahl, aktualisiert FROM dinge
		ORDER BY aktualisiert DESC
		LIMIT ?`

	tx, err := r.tm.BeginTx(ctx)
	if err != nil {
		return []Ding{}, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, statement, limit)
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
