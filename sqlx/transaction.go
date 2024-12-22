package sqlx

import (
	"context"
	"database/sql"
)

type Transaction interface {
	ExecContext(query string, args ...any) (sql.Result, error)
	QueryContext(query string, args ...any) (*sql.Rows, error)
	QueryRowContext(query string, args ...any) *sql.Row
	Commit() error
	Rollback() error
	Parent() Transaction
	Context() context.Context
}

type sqlTransaction struct {
	tm   *SqlTransactionManager
	conn *sql.Conn
	tx   *sql.Tx
	ctx  context.Context
}

func (t sqlTransaction) Parent() Transaction {
	return nil
}

func (t sqlTransaction) Commit() error {
	// TODO: Es fühlt sich so an, als gehöre dies zum TransactionManager
	// Es wird öfter auf die Felder des TransactionManagers zugegriffen,
	// als auf die Felder von t. Das gleiche gilt für Rollback.
	t.tm.mu.Lock()
	defer t.tm.mu.Unlock()

	ctx := t.ctx
	t.tm.tx[ctx] = t.Parent()
	return t.tx.Commit()
}

func (t *sqlTransaction) Rollback() error {
	t.tm.mu.Lock()
	defer t.tm.mu.Unlock()

	ctx := t.ctx
	conn := t.conn
	result := t.tx.Rollback()
	delete(t.tm.tx, ctx)
	if conn != nil {
		conn.Close()
	}

	return result
}

func (t sqlTransaction) ExecContext(query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(t.ctx, query, args...)
}

func (t sqlTransaction) QueryContext(query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, query, args...)
}

func (t sqlTransaction) QueryRowContext(query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, query, args...)
}

func (t sqlTransaction) Context() context.Context {
	return t.ctx
}

type nestedTransaction struct {
	tm       *SqlTransactionManager
	comitted bool
	parent   Transaction
}

func (t nestedTransaction) Context() context.Context {
	return t.parent.Context()
}

func (t *nestedTransaction) ExecContext(query string, args ...any) (sql.Result, error) {
	return t.parent.ExecContext(query, args...)
}

func (t *nestedTransaction) QueryContext(query string, args ...any) (*sql.Rows, error) {
	return t.parent.QueryContext(query, args...)
}

func (t *nestedTransaction) QueryRowContext(query string, args ...any) *sql.Row {
	return t.parent.QueryRowContext(query, args...)
}

func (t *nestedTransaction) Commit() error {
	ctx := t.Context()
	t.tm.tx[ctx] = t.parent
	t.comitted = true
	return nil
}

func (t nestedTransaction) Rollback() error {
	ctx := t.Context()
	t.tm.tx[ctx] = t.Parent()
	if !t.comitted {
		return t.parent.Rollback()
	}

	return nil
}

func (t nestedTransaction) Parent() Transaction {
	return t.parent
}
