package sqlx

import (
	"context"
	"database/sql"
)

type Transaction interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Commit() error
	Rollback() error
	Parent() Transaction
}

type SqlTransaction struct {
	tm *TransactionManager
	tx *sql.Tx
}

func (t SqlTransaction) Parent() Transaction {
	return nil
}

func (t SqlTransaction) Commit() error {
	t.tm.tx = t.Parent()
	return t.tx.Commit()
}

func (t SqlTransaction) Rollback() error {
	t.tm.tx = t.Parent()
	return t.tx.Rollback()
}

func (t SqlTransaction) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t SqlTransaction) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t SqlTransaction) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

type NestedTransaction struct {
	tm       *TransactionManager
	comitted bool
	parent   Transaction
}

func (t *NestedTransaction) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.parent.ExecContext(ctx, query, args...)
}

func (t *NestedTransaction) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.parent.QueryContext(ctx, query, args...)
}

func (t *NestedTransaction) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.parent.QueryRowContext(ctx, query, args...)
}

func (t *NestedTransaction) Commit() error {
	t.tm.tx = t.Parent()
	t.comitted = true
	return nil
}

func (t NestedTransaction) Rollback() error {
	t.tm.tx = t.Parent()
	if !t.comitted {
		return t.parent.Rollback()
	}

	return nil
}

func (t NestedTransaction) Parent() Transaction {
	return t.parent
}
