package ding_test

import (
	"context"
	"database/sql"
	"errors"
	"iter"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/haschi/dinge/ding"
	"github.com/haschi/dinge/sqlx"
	"github.com/haschi/dinge/system"
	"github.com/haschi/dinge/testx"
)

func TestRepository_GetById(t *testing.T) {

	type args struct {
		id  int64
		ctx context.Context
	}
	tests := []struct {
		name string
		// fields       repositoryProvider
		precondition testx.SetupFunc
		args         args
		want         ding.Ding
		wantErr      bool
	}{
		{
			name: "empty database",

			precondition: theFixture,
			args:         args{ctx: context.Background()},
			wantErr:      true,
		},
		{
			name: "lese ding 111",

			precondition: theFixture,
			args:         args{id: dinge[0].Id, ctx: context.Background()},
			want:         dinge[0],
			wantErr:      false,
		},
		{
			name: "closed database",

			precondition: closeDatabase,
			wantErr:      true,
		},
		{
			name: "context done",

			args:    args{ctx: newCanceledContext()},
			wantErr: true,
		},
		{
			name: "context nil",

			args:    args{ctx: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTransactionManager(t, tt.precondition, func(t *testing.T, tm sqlx.TransactionManager) {
				r := &ding.Repository{
					Clock: system.RealClock{},
					Tm:    tm,
				}

				got, err := r.GetById(tt.args.ctx, tt.args.id)
				if err != nil {
					if !tt.wantErr {
						t.Errorf("Repository.GetById() error = %v, wantErr %v", err, tt.wantErr)
					}
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Repository.GetById() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestRepository_MengeAktualisieren(t *testing.T) {

	type args struct {
		ctx   context.Context
		code  string
		menge int
	}
	tests := []struct {
		name string

		setup   testx.SetupFunc
		args    args
		want    ding.Ding
		wantErr bool
	}{
		{
			name:  "Update Paprika",
			args:  args{code: dinge[0].Code, menge: 42, ctx: context.Background()},
			setup: theFixture,
			want: ding.Ding{
				DingRef: ding.DingRef{
					Id:       dinge[0].Id,
					Code:     dinge[0].Code,
					Name:     dinge[0].Name,
					Anzahl:   dinge[0].Anzahl + 42,
					PhotoUrl: dinge[0].PhotoUrl,
				},
				Beschreibung: dinge[0].Beschreibung,
				Allgemein:    dinge[0].Allgemein,
				Aktualisiert: dinge[0].Aktualisiert,
			},
			wantErr: false,
		},
		{
			name: "Update unknown",

			setup:   theFixture,
			args:    args{code: "unknown", menge: 42, ctx: context.Background()},
			wantErr: true,
		},
		{
			name: "Update too much",

			setup:   theFixture,
			args:    args{code: dinge[0].Code, menge: -(dinge[0].Anzahl + 1), ctx: context.Background()},
			wantErr: true,
		},
		{
			name: "without context",

			args:    args{code: dinge[0].Code, menge: 1, ctx: nil},
			wantErr: true,
		},
		{
			name: "with canceled context",

			setup:   testx.SetupFunc(theFixture).AndThen(closeDatabase),
			args:    args{code: "", menge: 0, ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			withTransactionManager(t, tt.setup, func(t *testing.T, tm sqlx.TransactionManager) {
				r := &ding.Repository{
					Clock: FixedClock{Timestamp: tt.want.Aktualisiert},
					Tm:    tm,
				}
				ding, err := r.MengeAktualisieren(tt.args.ctx, tt.args.code, tt.args.menge)

				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.MengeAktualisieren() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if err != nil {
					return
				}

				got, err := r.GetById(tt.args.ctx, ding.Id)
				if err != nil {
					t.Fatal("ding should exists", err)
				}

				if got != tt.want {
					t.Errorf("Repository.MengeAktualisieren() = %v, want %v", got, tt.want)
				}

			})
		})
	}
}

func TestRepository_Insert(t *testing.T) {

	type args struct {
		ctx    context.Context
		code   string
		anzahl int
	}
	tests := []struct {
		name string

		precondition testx.SetupFunc
		args         args
		want         ding.InsertResult
		wantErr      bool
	}{
		{
			name: "closed database",

			precondition: testx.SetupFunc(theFixture).AndThen(closeDatabase),
			args:         args{ctx: context.Background(), code: "doesn't matter", anzahl: 42},
			wantErr:      true,
		},
		{
			name: "Insert new ding",

			precondition: theFixture,
			args:         args{ctx: context.Background(), code: "QWERT", anzahl: 1},
			want:         ding.InsertResult{Id: int64(len(dinge) + 1), Created: true},
		},
		{
			name: "Insert existing ding",

			precondition: theFixture,
			args:         args{ctx: context.Background(), code: dinge[0].Code, anzahl: 1},
			want:         ding.InsertResult{Id: dinge[0].Id, Created: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTransactionManager(t, tt.precondition, func(t *testing.T, tm sqlx.TransactionManager) {
				r := ding.Repository{
					Clock: system.RealClock{},
					Tm:    tm,
				}

				got, err := r.Insert(tt.args.ctx, tt.args.code, tt.args.anzahl)
				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.Insert() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Repository.Insert() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestRepository_NamenAktualisieren(t *testing.T) {

	type args struct {
		ctx     context.Context
		anfrage ding.Aktualisierungsanfrage
	}
	tests := []struct {
		name         string
		timestamp    time.Time
		precondition testx.SetupFunc
		args         args
		want         ding.Ding
		wantErr      bool
	}{
		{
			name:         "closed database",
			precondition: closeDatabase,
			args:         args{ctx: context.Background()},
			wantErr:      true,
		},
		{
			name:    "without context",
			args:    args{ctx: nil},
			wantErr: true,
		},
		{
			name:         "Beschreibung ändern",
			precondition: theFixture,
			args: args{
				ctx: context.Background(),
				anfrage: ding.Aktualisierungsanfrage{
					Id:           dinge[0].Id,
					Name:         "Pepperoni",
					Beschreibung: "Neue Beschreibung",
					Allgemein:    dinge[0].Allgemein,
				},
			},
			timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			want: ding.Ding{
				DingRef: ding.DingRef{
					Id:       dinge[0].Id,
					Name:     "Pepperoni",
					Code:     dinge[0].Code,
					Anzahl:   dinge[0].Anzahl,
					PhotoUrl: dinge[0].PhotoUrl,
				},
				Beschreibung: "Neue Beschreibung",
				Allgemein:    dinge[0].Allgemein,
				Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			},
			wantErr: false,
		},
		{
			name: "change Paprika",

			precondition: theFixture,
			args: args{
				ctx: context.Background(),
				anfrage: ding.Aktualisierungsanfrage{
					Id:           dinge[0].Id,
					Name:         "Pepperoni",
					Beschreibung: dinge[0].Beschreibung,
					Allgemein:    dinge[0].Allgemein,
				},
			},
			timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			want: ding.Ding{
				DingRef: ding.DingRef{
					Id:       dinge[0].Id,
					Name:     "Pepperoni",
					Code:     dinge[0].Code,
					Anzahl:   dinge[0].Anzahl,
					PhotoUrl: dinge[0].PhotoUrl,
				},
				Beschreibung: dinge[0].Beschreibung,
				Allgemein:    dinge[0].Allgemein,
				Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			},
			wantErr: false,
		},
		{
			name: "update unknown",

			precondition: theFixture,
			args: args{ctx: context.Background(),
				anfrage: ding.Aktualisierungsanfrage{
					Id:   42,
					Name: "doesn't matter",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTransactionManager(t, tt.precondition, func(t *testing.T, tm sqlx.TransactionManager) {
				r := &ding.Repository{
					Clock: FixedClock{Timestamp: tt.timestamp},
					Tm:    tm,
				}

				if err := r.Aktualisieren(tt.args.ctx, tt.args.anfrage); (err != nil) != tt.wantErr {
					t.Errorf("Repository.NamenAktualisieren() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					got, err := r.GetById(tt.args.ctx, tt.args.anfrage.Id)
					if err != nil {
						t.Fatal("unexpected condition")
					}
					if !reflect.DeepEqual(got, tt.want) {
						t.Errorf("Repository.GetById() = %v, want %v", got, tt.want)
					}
				}
			})
		})
	}
}

func TestRepository_GetLatest(t *testing.T) {

	type args struct {
		limit int
		query string
		sort  string
	}

	tests := []struct {
		name string

		precondition testx.SetupFunc
		args         args
		want         []ding.DingRef
		wantErr      bool
	}{
		{
			name: "empty database",

			precondition: theFixture,
			want:         []ding.DingRef{},
		},
		{
			name: "database closed",

			precondition: closeDatabase,
			wantErr:      true,
		},
		{
			name: "limit 0",

			precondition: theFixture,
			args:         args{limit: 0},
			want:         []ding.DingRef{},
		},
		{
			name: "limit is less than the number of stored items",

			precondition: theFixture,
			args:         args{limit: 2},
			want:         mapToDingeRef([]ding.Ding{dinge[2], dinge[1]}),
		},
		{
			name: "limit is greater than the number of stored items",

			precondition: theFixture,
			args:         args{limit: 4},
			want:         mapToDingeRef([]ding.Ding{dinge[2], dinge[1], dinge[0]}),
		},
		{
			name: "items are sorted by date of change",

			precondition: testx.SetupFunc(theFixture).AndThen(func(d *sql.DB) error {
				clock := FixedClock{Timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:58:05"))}
				tm, err := sqlx.NewSqlTransactionManager(d)
				if err != nil {
					t.Fatal(err)
				}
				repository := &ding.Repository{
					Clock: clock,
					Tm:    tm,
				}

				if err != nil {
					return err
				}
				_, err = repository.MengeAktualisieren(context.Background(), dinge[1].Code, 1)
				return err
			}),

			args: args{limit: 4},
			want: mapToDingeRef([]ding.Ding{
				{
					DingRef: ding.DingRef{
						Id:       dinge[1].Id,
						Name:     dinge[1].Name,
						Code:     dinge[1].Code,
						Anzahl:   dinge[1].Anzahl + 1,
						PhotoUrl: dinge[1].PhotoUrl,
					},
					Beschreibung: dinge[1].Beschreibung,
				},
				dinge[2],
				dinge[0],
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTransactionManager(t, tt.precondition, func(t *testing.T, tm sqlx.TransactionManager) {
				r := &ding.Repository{
					Clock: system.RealClock{},
					Tm:    tm,
				}

				got, err := r.Search(context.Background(), tt.args.limit, tt.args.query, tt.args.sort)
				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.GetLatest() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Repository.GetLatest() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestRepository_ForeignKeys(t *testing.T) {
	withDatabase(t, theFixture, func(t *testing.T, db *sql.DB) {
		row := db.QueryRow("PRAGMA foreign_keys")
		var result int
		if err := row.Scan(&result); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				t.Fatal("Fremdschlüssel werde nicht von der Datenbank unterstützt.", err)
			}
			t.Fatal(err)
		}

		if result == 0 {
			t.Error("Fremdschlüssel sind deaktiviert.")
		}
	})
}

func transform[T, U any](s iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range s {
			ref := fn(v)
			if !yield(ref) {
				return
			}
		}
	}
}

func mapToDingeRef(d []ding.Ding) []ding.DingRef {
	mapped := transform(slices.Values(d), func(d ding.Ding) ding.DingRef {
		return d.DingRef
	})

	return slices.Collect(mapped)
}

// Helfer, die möglicherweise in ein eigenes Package gehören.

// FixedClock ist eine Implementierung von [model.Clock] für Testzwecke
type FixedClock struct {
	Timestamp time.Time
}

// Now liefert stets den in [FixedClock.Timestamp] festgelegten Zeitstempel zurück.
func (c FixedClock) Now() time.Time {
	return c.Timestamp
}

// must stoppt die Ausführung, wenn err nicht nil ist. Ansonsten liefert die Funktion den Wert value zurück.
//
// Die Funktion sollte nur in Testcode verwendet werden.
func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

func newCanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return ctx
}

func closeDatabase(db *sql.DB) error {
	return db.Close()
}

// testFunc deklariert eine Testfunktion, die eine Datenbank benutzt
type testFunc func(*testing.T, *sql.DB)

// thefixture liefert eine Funktion, mit der die Vorbedingung eines Tests hergestellt werden kann
//
// Die zurückgegebene Funktion initialisiert die Datenbank und füllt diese mit Testdaten, wenn sie aufgerufen wird.
//
// Die zurückgegebene Funktion kann mit [setup.AndThen] mit einer weiteren Funktion kombiniert werden.
func theFixture(db *sql.DB) error {
	return sqlx.ExecuteScripts(db, ding.CreateScript, ding.FixtureScript)
}

// withDatabase stellt eine Ausführungsumgebung bereit, in der eine Testfunktion mit Datenbank ausgeführt werden kann.
func withDatabase(t *testing.T, setupFn testx.SetupFunc, testFn testFunc) {
	t.Helper()

	db, err := sqlx.NewTestDatabase()
	if err != nil {
		t.Fatal("can not open database", err)
	}

	db.SetMaxOpenConns(0)
	defer db.Close()

	if setupFn != nil {
		if err := setupFn(db); err != nil {
			t.Fatal("can not setup fixture", err)
		}
	}

	if testFn == nil {
		t.Fatal("no test function provided")
	}

	testFn(t, db)
}

// Testdaten. Korrespondieren mit model/fixture.sql
var dinge = []ding.Ding{
	{
		DingRef: ding.DingRef{
			Id:       1,
			Name:     "Paprika",
			Code:     "111",
			Anzahl:   1,
			PhotoUrl: "/photos/1",
		},
		Beschreibung: "Eine Planzengattung, die zur Familie der Nachtschattengewächse gehört",
		Allgemein:    "Gemüse",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 18:48:01")),
	},

	{
		DingRef: ding.DingRef{
			Id:       2,
			Name:     "Gurke",
			Code:     "222",
			Anzahl:   2,
			PhotoUrl: "/photos/2",
		},
		Beschreibung: "",
		Allgemein:    "Gemüse",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:05:02")),
	},

	{
		DingRef: ding.DingRef{
			Id:       3,
			Name:     "Tomate",
			Code:     "333",
			Anzahl:   3,
			PhotoUrl: "/photos/3",
		},
		Beschreibung: "",
		Allgemein:    "Gemüse",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:06:03")),
	},
}

func withTransactionManager(t *testing.T, setupFn func(*sql.DB) error, testFn func(*testing.T, sqlx.TransactionManager)) {
	t.Helper()
	db, err := sqlx.NewTestDatabase()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()
	db.SetMaxOpenConns(0)

	tm, err := sqlx.NewSqlTransactionManager(db)
	if err != nil {
		t.Fatal(err)
	}

	if setupFn != nil {
		if err := setupFn(db); err != nil {
			t.Fatal("can not setup fixture", err)
		}
	}

	if testFn == nil {
		t.Fatal("no test function profided")
	}

	testFn(t, tm)
}
