package model_test

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/system"
	"github.com/haschi/dinge/testx"
	_ "github.com/mattn/go-sqlite3"
)

func TestRepository_GetById(t *testing.T) {

	type args struct {
		id  int64
		ctx context.Context
	}
	tests := []struct {
		name         string
		fields       repositoryProvider
		precondition testx.SetupFunc
		args         args
		want         model.Ding
		wantErr      bool
	}{
		{
			name:         "empty database",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background()},
			wantErr:      true,
		},
		{
			name:         "lese ding 111",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			args:         args{id: dinge[0].Id, ctx: context.Background()},
			want:         dinge[0],
			wantErr:      false,
		},
		{
			name:         "closed database",
			fields:       model.NewRepository,
			precondition: closeDatabase,
			wantErr:      true,
		},
		{
			name:    "context done",
			fields:  model.NewRepository,
			args:    args{ctx: newCanceledContext()},
			wantErr: true,
		},
		{
			name:    "context nil",
			fields:  model.NewRepository,
			args:    args{ctx: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r, err := tt.fields(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
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

func TestRepository_GetByCode(t *testing.T) {

	withDatabase(t, thefixture(dinge), func(t *testing.T, db *sql.DB) {
		type args struct {
			ctx  context.Context
			code string
		}
		tests := []struct {
			name    string
			fields  repositoryProvider
			args    args
			want    model.Ding
			wantErr bool
		}{
			{
				name:    "Get existing ding",
				fields:  model.NewRepository,
				args:    args{ctx: context.Background(), code: dinge[0].Code},
				want:    dinge[0],
				wantErr: false,
			},
			{
				name:    "Get unknown ding",
				fields:  model.NewRepository,
				args:    args{ctx: context.Background(), code: "unknown"},
				wantErr: true,
			},
			{
				name:    "without context",
				fields:  model.NewRepository,
				args:    args{},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r, err := tt.fields(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
				}
				got, err := r.GetByCode(tt.args.ctx, tt.args.code)
				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.GetByCode() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Repository.GetByCode() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestRepository_MengeAktualisieren(t *testing.T) {

	type args struct {
		ctx   context.Context
		code  string
		menge int
	}
	tests := []struct {
		name    string
		fields  repositoryProvider
		setup   testx.SetupFunc
		args    args
		want    model.Ding
		wantErr bool
	}{
		{
			name:   "Update Paprika",
			fields: model.NewRepository,
			args:   args{code: dinge[0].Code, menge: 42, ctx: context.Background()},
			setup:  thefixture(dinge),
			want: model.Ding{
				Id:     dinge[0].Id,
				Code:   dinge[0].Code,
				Name:   dinge[0].Name,
				Anzahl: dinge[0].Anzahl + 42,
			},
			wantErr: false,
		},
		{
			name:    "Update unknown",
			fields:  model.NewRepository,
			setup:   thefixture(dinge),
			args:    args{code: "unknown", menge: 42, ctx: context.Background()},
			wantErr: true,
		},
		{
			name:    "Update too much",
			fields:  model.NewRepository,
			setup:   thefixture(dinge),
			args:    args{code: dinge[0].Code, menge: -(dinge[0].Anzahl + 1), ctx: context.Background()},
			wantErr: true,
		},
		{
			name:    "without context",
			fields:  model.NewRepository,
			args:    args{code: dinge[0].Code, menge: 1, ctx: nil},
			wantErr: true,
		},
		{
			name:    "with canceled context",
			fields:  model.NewRepository,
			setup:   thefixture(dinge).AndThen(closeDatabase),
			args:    args{code: "", menge: 0, ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			withDatabase(t, tt.setup, func(t *testing.T, db *sql.DB) {
				r, err := tt.fields(db, FixedClock{Timestamp: tt.want.Aktualisiert})
				if err != nil {
					t.Fatal(err)
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
		name         string
		fields       repositoryProvider
		precondition testx.SetupFunc
		args         args
		want         model.InsertResult
		wantErr      bool
	}{
		{
			name:         "closed database",
			fields:       model.NewRepository,
			precondition: thefixture(dinge).AndThen(closeDatabase),
			args:         args{ctx: context.Background(), code: "doesn't matter", anzahl: 42},
			wantErr:      true,
		},
		{
			name:         "Insert new ding",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background(), code: "QWERT", anzahl: 1},
			want:         model.InsertResult{Id: int64(len(dinge) + 1), Created: true},
		},
		{
			name:         "Insert existing ding",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background(), code: dinge[0].Code, anzahl: 1},
			want:         model.InsertResult{Id: dinge[0].Id, Created: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r, err := tt.fields(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
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
		id        int64
		name      string
		timestamp time.Time
		ctx       context.Context
	}
	tests := []struct {
		name         string
		fields       repositoryProvider
		precondition testx.SetupFunc
		args         args
		want         model.Ding
		wantErr      bool
	}{
		{
			name:         "closed database",
			precondition: closeDatabase,
			fields:       model.NewRepository,
			args:         args{ctx: context.Background()},
			wantErr:      true,
		},
		{
			name:    "without context",
			fields:  model.NewRepository,
			args:    args{ctx: nil},
			wantErr: true,
		},
		{
			name:         "change Paprika",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			args: args{
				ctx:       context.Background(),
				id:        dinge[0].Id,
				name:      "Pepperoni",
				timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			},
			want: model.Ding{
				Id:           dinge[0].Id,
				Name:         "Pepperoni",
				Code:         dinge[0].Code,
				Anzahl:       dinge[0].Anzahl,
				Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			},
			wantErr: false,
		},
		{
			name:         "update unknown",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background(), id: 42, name: "doesn't matter"},
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r, err := tt.fields(db, FixedClock{Timestamp: tt.args.timestamp})
				if err != nil {
					t.Fatal(err)
				}

				if err := r.NamenAktualisieren(tt.args.ctx, tt.args.id, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("Repository.NamenAktualisieren() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					got, err := r.GetById(tt.args.ctx, tt.args.id)
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

	type fields func(*sql.DB, model.Clock) (*model.Repository, error)

	tests := []struct {
		name         string
		fields       fields
		precondition testx.SetupFunc
		arg          int
		want         []model.Ding
		wantErr      bool
	}{
		{
			name:         "empty database",
			fields:       model.NewRepository,
			precondition: thefixture([]model.Ding{}),
			want:         []model.Ding{},
		},
		{
			name:         "database closed",
			fields:       model.NewRepository,
			precondition: closeDatabase,
			wantErr:      true,
		},
		{
			name:         "limit 0",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			arg:          0,
			want:         []model.Ding{},
		},
		{
			name:         "limit is less than the number of stored items",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			arg:          2,
			want:         []model.Ding{dinge[2], dinge[1]},
		},
		{
			name:         "limit is greater than the number of stored items",
			fields:       model.NewRepository,
			precondition: thefixture(dinge),
			arg:          4,
			want:         []model.Ding{dinge[2], dinge[1], dinge[0]},
		},
		{
			name:   "items are sorted by date of change",
			fields: model.NewRepository,
			precondition: thefixture(dinge).AndThen(func(d *sql.DB) error {
				clock := FixedClock{Timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:58:05"))}
				repository, err := model.NewRepository(d, clock)

				if err != nil {
					return err
				}
				_, err = repository.MengeAktualisieren(context.Background(), dinge[1].Code, 1)
				return err
			}),

			arg: 4,
			want: []model.Ding{
				{
					Id:           dinge[1].Id,
					Name:         dinge[1].Name,
					Code:         dinge[1].Code,
					Anzahl:       dinge[1].Anzahl + 1,
					Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:58:05")),
				},
				dinge[2],
				dinge[0],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r, err := tt.fields(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
				}

				got, err := r.GetLatest(context.Background(), tt.arg)
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

// Helfer, die möglicherweise in ein eigenes Package gehören.

// FixedClock ist eine Implementierung von [model.Clock] für Testzwecke
type FixedClock struct {
	Timestamp time.Time
}

// Now liefert stets den in [FixedClock.Timestamp] festgelegten Zeitstempel zurück.
func (c FixedClock) Now() time.Time {
	return c.Timestamp
}

// repositoryProvider ist eine Funktion, die ein [model.Repository] herstellt.
type repositoryProvider func(*sql.DB, model.Clock) (*model.Repository, error)

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

func TestNewRepository(t *testing.T) {
	type args struct {
		dbfunc func(*sql.DB) *sql.DB
		clock  model.Clock
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
		wantErr bool
	}{
		{
			name:    "no database",
			args:    args{dbfunc: func(d *sql.DB) *sql.DB { return nil }, clock: system.RealClock{}},
			wantNil: true,
			wantErr: true,
		},
		{
			name:    "with database",
			args:    args{dbfunc: func(d *sql.DB) *sql.DB { return d }, clock: system.RealClock{}},
			wantNil: false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, nil, func(t *testing.T, d *sql.DB) {
				got, err := model.NewRepository(tt.args.dbfunc(d), tt.args.clock)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewRepository() error = %v, wantErr %v", err, tt.wantErr)
				}

				if (got == nil) != tt.wantNil {
					t.Errorf("NewRepository() repository = %v, wantNil %v", got, tt.wantNil)
				}
			})
		})
	}
}

func closeDatabase(db *sql.DB) error {
	return db.Close()
}

const dataSource = "file::memory:?cache=shared"

func setupFixture(ctx context.Context, dinge []model.Ding) testx.SetupFunc {

	return func(db *sql.DB) error {
		for _, ding := range dinge {

			r, err := model.NewRepository(db, FixedClock{Timestamp: ding.Aktualisiert})
			if err != nil {
				return err
			}

			result, err := r.Insert(ctx, ding.Code, ding.Anzahl)
			if err != nil {
				return err
			}

			if !result.Created {
				return errors.New("ding bereits vorhanden")
			}

			if result.Id != ding.Id {
				return errors.New("id missmatch")
			}

			if err = r.NamenAktualisieren(ctx, result.Id, ding.Name); err != nil {
				return err
			}
		}

		// Nur zu Testzwecken

		// rows, err := db.Query("SELECT id, name, code, anzahl, aktualisiert FROM dinge")
		// if err != nil {
		// 	return err
		// }

		// defer rows.Close()

		// for rows.Next() {
		// 	var ding model.Ding
		// 	err = rows.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	fmt.Println(ding)
		// }

		return nil
	}
}

// testFunc deklariert eine Testfunktion, die eine Datenbank benutzt
type testFunc func(*testing.T, *sql.DB)

// thefixture liefert eine Funktion, mit der die Vorbedingung eines Tests hergestellt werden kann
//
// Die zurückgegebene Funktion initialisiert die Datenbank und füllt diese mit Testdaten, wenn sie aufgerufen wird.
//
// Die zurückgegebene Funktion kann mit [setup.AndThen] mit einer weiteren Funktion kombiniert werden.
func thefixture(dinge []model.Ding) testx.SetupFunc {
	return testx.SetupFunc(model.CreateTable).AndThen(setupFixture(context.Background(), dinge))
}

// withDatabase stellt eine Ausführungsumgebung bereit, in der eine Testfunktion mit Datenbank ausgeführt werden kann.
func withDatabase(t *testing.T, setupFn testx.SetupFunc, testFn testFunc) {
	t.Helper()

	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		t.Fatal("can not open database", err)
	}

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
var dinge = []model.Ding{
	{Id: 1, Name: "Paprika", Code: "111", Anzahl: 1, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 18:48:01"))},

	{Id: 2, Name: "Gurke", Code: "222", Anzahl: 2, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:05:02"))},

	{Id: 3, Name: "Tomate", Code: "333", Anzahl: 3, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:06:03"))},
}
