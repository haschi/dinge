package sqlx

import (
	"context"
	"database/sql"
)

type TransactionManager struct {
	db *sql.DB
	tx Transaction
}

func (tm *TransactionManager) BeginTx(ctx context.Context) (Transaction, error) {
	if tm.tx == nil {
		tx, err := tm.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		tm.tx = SqlTransaction{tx: tx, tm: tm}

		return tm.tx, nil
	}

	nested := &NestedTransaction{parent: tm.tx, comitted: false, tm: tm}
	tm.tx = nested
	return nested, nil
}
