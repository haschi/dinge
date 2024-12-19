package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

type TransactionManager interface {
	BeginTx(ctx context.Context) (Transaction, error)
	Count() int
}

type SqlTransactionManager struct {
	mu sync.Mutex
	db *sql.DB
	tx map[context.Context]Transaction
}

func NewSqlTransactionManager(db *sql.DB) (*SqlTransactionManager, error) {
	if db == nil {
		return nil, errors.New("no database provided")
	}

	tm := &SqlTransactionManager{
		db: db,
		tx: make(map[context.Context]Transaction),
	}

	return tm, nil
}

// BeginTx startet eine Transaktion
//
// Für den ersten Aufruf von BeginTx für den [context.Context] startet der Transaktionsmanager eine Datenbank-Transaktion. Für jeden weiteren Aufruf wird eine geschachtelte Transaktion erstellt, die mit der ursprünglichen Datenbank-Transaktion verkettet ist. Ein Rollback der geschachtelten Transaktion führ zu einem Rollback der Datenbank-Transaktion. Ein Commit der geschachtelten Transaktion hat keine Auswirkung auf die Datenbanktransaktion. Ein Rollback oder Commit einer beliebigen Transaktion, löst die Transaktion vom Transaktionmanager und übergibt die Kontrolle an die jeweils übergeordnete Transaktion zurück.
func (tm *SqlTransactionManager) BeginTx(ctx context.Context) (Transaction, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, ok := tm.tx[ctx]

	if !ok {
		var conn *sql.Conn
		conn, err := tm.db.Conn(ctx)
		if err != nil {
			return nil, err
		}

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}

		tm.tx[ctx] = &sqlTransaction{tx: tx, conn: conn, tm: tm, ctx: ctx}
		return tm.tx[ctx], nil
	}

	nested := &nestedTransaction{parent: tx, comitted: false, tm: tm}
	tm.tx[ctx] = nested
	return nested, nil
}

func (tm *SqlTransactionManager) GetDb() *sql.DB {
	return tm.db
}

// Count lifert die Anzahl offener Transaktionen
func (tm *SqlTransactionManager) Count() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return len(tm.tx)
}
