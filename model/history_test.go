package model_test

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/sqlx"
	"github.com/haschi/dinge/system"
	"github.com/haschi/dinge/testx"
)

func TestRepository_LogEvent(t *testing.T) {

	type args struct {
		ctx       context.Context
		operation int
		count     int
		dingId    int64
	}

	validEvent := args{
		ctx:       context.Background(),
		operation: 1,
		count:     1,
		dingId:    dinge[0].Id,
	}

	invalidDing := args{
		ctx:       context.Background(),
		operation: 1,
		count:     2,
		dingId:    666,
	}

	invalidOperation := args{
		ctx:       context.Background(),
		operation: 666,
		count:     2,
		dingId:    dinge[0].Id,
	}

	tests := []struct {
		name         string
		precondition testx.SetupFunc
		args         args
		maxTxOps     int // -1 für beliebig viele
		wantErr      error
	}{
		{
			name:         "insert new event for existing ding",
			precondition: theFixture,
			args:         validEvent,
			maxTxOps:     -1,
			wantErr:      nil,
		},
		{
			name:         "closed database",
			precondition: testx.SetupFunc(theFixture).AndThen(closeDatabase),
			args:         validEvent,
			maxTxOps:     -1,
			wantErr:      model.ErrDataAccess,
		},
		{
			name:         "context closed before transaction",
			precondition: theFixture,
			args:         validEvent,
			maxTxOps:     0,
			wantErr:      model.ErrDataAccess,
		},
		{
			name:         "context closed during transaction",
			precondition: theFixture,
			args:         validEvent,
			maxTxOps:     1,
			wantErr:      model.ErrDataAccess,
		},
		{
			name:         "context closed before commit",
			precondition: theFixture,
			args:         validEvent,
			maxTxOps:     2,
			wantErr:      model.ErrDataAccess,
		},
		{
			name:         "foreign key violation ding",
			precondition: theFixture,
			args:         invalidDing,
			maxTxOps:     -1,
			wantErr:      model.ErrInvalidParameter,
		},
		{
			name:         "foreign key violation operation",
			precondition: theFixture,
			args:         invalidOperation,
			maxTxOps:     -1,
			wantErr:      model.ErrInvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				repository, err := model.NewRepository(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
				}

				repository.Tm = must(NewCancableTransactionManager(db, tt.maxTxOps))

				if err := repository.LogEvent(tt.args.ctx, tt.args.operation, tt.args.count, tt.args.dingId); !errors.Is(err, tt.wantErr) {
					t.Errorf("Repository.InsertEvent() error = %v, want %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestRepository_GetHistory(t *testing.T) {

	fullHistory := []model.Event{
		{
			Operation: 1,
			Anzahl:    3,
			DingRef: model.DingRef{
				Id:   3,
				Name: "Tomate",
				Code: "333",
			},
		},
		{
			Operation: 1,
			Anzahl:    2,
			DingRef: model.DingRef{
				Id:   2,
				Name: "Gurke",
				Code: "222",
			},
		},
		{

			Operation: 1,
			Anzahl:    1,
			DingRef: model.DingRef{
				Id:   1,
				Name: "Paprika",
				Code: "111",
			},
		},
	}

	type args struct {
		limit int
	}
	tests := []struct {
		name    string
		args    args
		want    []model.Event
		wantErr bool
	}{
		{
			name: "limit greater then elements",
			args: args{limit: len(fullHistory) + 1},
			want: fullHistory,
		},
		{
			name: "limit 1",
			args: args{limit: 1},
			want: fullHistory[:1],
		},
		{
			name: "limit 2",
			args: args{limit: 2},
			want: fullHistory[:2],
		},
		{
			name: "limit 3",
			args: args{limit: 3},
			want: fullHistory[:],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, theFixture, func(t *testing.T, db *sql.DB) {
				repository, err := model.NewRepository(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
				}

				got, err := repository.GetHistory(context.Background(), tt.args.limit)
				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.GetHistory() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !slices.Equal(got, tt.want) {
					t.Errorf("Repository.GetHistory() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

// CancelableTransactionManager ist ein [sqlx.TransactionManager], der seinen [context.Context] nach einer vorgegebenen Anzahl von Operationen abschließt (canceled).
//
// Dieser TransactionManager wird in Unit Tests verwendet, um Fehler zu simulieren.
type CancelableTransactionManager struct {
	*sqlx.SqlTransactionManager
	count int
}

func NewCancableTransactionManager(db *sql.DB, count int) (*CancelableTransactionManager, error) {

	tm, err := sqlx.NewSqlTransactionManager(db)
	if err != nil {
		return nil, err
	}

	return &CancelableTransactionManager{
		SqlTransactionManager: tm,
		count:                 count,
	}, nil
}

func (tm *CancelableTransactionManager) BeginTx(ctx context.Context) (sqlx.Transaction, error) {
	cancable, cancelFn := context.WithCancel(ctx)
	if tm.count == 0 {
		cancelFn()
	}

	tx, err := tm.SqlTransactionManager.BeginTx(cancable)
	ct := &CancelableTransaction{
		Transaction: tx,
		count:       tm.count - 1,
		cancel:      cancelFn,
	}
	return ct, err
}

type CancelableTransaction struct {
	sqlx.Transaction
	count  int
	cancel context.CancelFunc
}

func (t *CancelableTransaction) cancelAndDecrease() {
	if t.count == 0 {
		t.cancel()
	}
	t.count = t.count - 1
}

func (t *CancelableTransaction) ExecContext(query string, args ...any) (sql.Result, error) {
	t.cancelAndDecrease()
	return t.Transaction.ExecContext(query, args...)
}

func (t *CancelableTransaction) QueryContext(query string, args ...any) (*sql.Rows, error) {
	t.cancelAndDecrease()
	return t.Transaction.QueryContext(query, args...)
}

func (t *CancelableTransaction) QueryRowContext(query string, args ...any) *sql.Row {
	t.cancelAndDecrease()
	return t.Transaction.QueryRowContext(query, args...)
}

func (t *CancelableTransaction) Commit() error {
	t.cancelAndDecrease()
	return t.Transaction.Commit()
}

func (t *CancelableTransaction) Rollback() error {
	defer t.cancel()
	return t.Transaction.Rollback()
}
