package model_test

import (
	"context"
	"database/sql"
	_ "embed"
	"iter"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/system"
	"github.com/haschi/dinge/testx"
	_ "github.com/mattn/go-sqlite3"
)

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
			precondition: theFixture,
			args:         args{ctx: context.Background()},
			wantErr:      true,
		},
		{
			name:         "lese ding 111",
			fields:       model.NewRepository,
			precondition: theFixture,
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

	withDatabase(t, theFixture, func(t *testing.T, db *sql.DB) {
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
			setup:  theFixture,
			want: model.Ding{
				DingRef: model.DingRef{
					Id:     dinge[0].Id,
					Code:   dinge[0].Code,
					Name:   dinge[0].Name,
					Anzahl: dinge[0].Anzahl + 42,
				},
				Beschreibung: dinge[0].Beschreibung,
				Aktualisiert: dinge[0].Aktualisiert,
			},
			wantErr: false,
		},
		{
			name:    "Update unknown",
			fields:  model.NewRepository,
			setup:   theFixture,
			args:    args{code: "unknown", menge: 42, ctx: context.Background()},
			wantErr: true,
		},
		{
			name:    "Update too much",
			fields:  model.NewRepository,
			setup:   theFixture,
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
			setup:   testx.SetupFunc(theFixture).AndThen(closeDatabase),
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
			precondition: testx.SetupFunc(theFixture).AndThen(closeDatabase),
			args:         args{ctx: context.Background(), code: "doesn't matter", anzahl: 42},
			wantErr:      true,
		},
		{
			name:         "Insert new ding",
			fields:       model.NewRepository,
			precondition: theFixture,
			args:         args{ctx: context.Background(), code: "QWERT", anzahl: 1},
			want:         model.InsertResult{Id: int64(len(dinge) + 1), Created: true},
		},
		{
			name:         "Insert existing ding",
			fields:       model.NewRepository,
			precondition: theFixture,
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
		ctx          context.Context
		id           int64
		name         string
		beschreibung string
		//timestamp    time.Time
	}
	tests := []struct {
		name         string
		fields       repositoryProvider
		timestamp    time.Time
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
			name:         "Beschreibung ändern",
			fields:       model.NewRepository,
			precondition: theFixture,
			args: args{
				ctx:          context.Background(),
				id:           dinge[0].Id,
				name:         "Pepperoni",
				beschreibung: "Neue Beschreibung",
			},
			timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			want: model.Ding{
				DingRef: model.DingRef{
					Id:     dinge[0].Id,
					Name:   "Pepperoni",
					Code:   dinge[0].Code,
					Anzahl: dinge[0].Anzahl,
				},
				Beschreibung: "Neue Beschreibung",
				Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			},
			wantErr: false,
		},
		{
			name:         "change Paprika",
			fields:       model.NewRepository,
			precondition: theFixture,
			args: args{
				ctx:          context.Background(),
				id:           dinge[0].Id,
				name:         "Pepperoni",
				beschreibung: dinge[0].Beschreibung,
			},
			timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			want: model.Ding{
				DingRef: model.DingRef{
					Id:     dinge[0].Id,
					Name:   "Pepperoni",
					Code:   dinge[0].Code,
					Anzahl: dinge[0].Anzahl,
				},
				Beschreibung: dinge[0].Beschreibung,
				Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:38:04")),
			},
			wantErr: false,
		},
		{
			name:         "update unknown",
			fields:       model.NewRepository,
			precondition: theFixture,
			args:         args{ctx: context.Background(), id: 42, name: "doesn't matter"},
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r, err := tt.fields(db, FixedClock{Timestamp: tt.timestamp})
				if err != nil {
					t.Fatal(err)
				}

				if err := r.DingAktualisieren(tt.args.ctx, tt.args.id, tt.args.name, tt.args.beschreibung); (err != nil) != tt.wantErr {
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
		want         []model.DingRef
		wantErr      bool
	}{
		{
			name:         "empty database",
			fields:       model.NewRepository,
			precondition: theFixture,
			want:         []model.DingRef{},
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
			precondition: theFixture,
			arg:          0,
			want:         []model.DingRef{},
		},
		{
			name:         "limit is less than the number of stored items",
			fields:       model.NewRepository,
			precondition: theFixture,
			arg:          2,
			want:         mapToDingeRef([]model.Ding{dinge[2], dinge[1]}),
		},
		{
			name:         "limit is greater than the number of stored items",
			fields:       model.NewRepository,
			precondition: theFixture,
			arg:          4,
			want:         mapToDingeRef([]model.Ding{dinge[2], dinge[1], dinge[0]}),
		},
		{
			name:   "items are sorted by date of change",
			fields: model.NewRepository,
			precondition: testx.SetupFunc(theFixture).AndThen(func(d *sql.DB) error {
				clock := FixedClock{Timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:58:05"))}
				repository, err := model.NewRepository(d, clock)

				if err != nil {
					return err
				}
				_, err = repository.MengeAktualisieren(context.Background(), dinge[1].Code, 1)
				return err
			}),

			arg: 4,
			want: mapToDingeRef([]model.Ding{
				{
					DingRef: model.DingRef{
						Id:     dinge[1].Id,
						Name:   dinge[1].Name,
						Code:   dinge[1].Code,
						Anzahl: dinge[1].Anzahl + 1,
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

// TODO: Nur vorläufig. Das muss besser herausgearbeitet werden
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

func mapToDingeRef(d []model.Ding) []model.DingRef {
	mapped := transform(slices.Values(d), func(d model.Ding) model.DingRef {
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

func closeDatabase(db *sql.DB) error {
	return db.Close()
}

const dataSource = "file::memory:?cache=shared"

// testFunc deklariert eine Testfunktion, die eine Datenbank benutzt
type testFunc func(*testing.T, *sql.DB)

// thefixture liefert eine Funktion, mit der die Vorbedingung eines Tests hergestellt werden kann
//
// Die zurückgegebene Funktion initialisiert die Datenbank und füllt diese mit Testdaten, wenn sie aufgerufen wird.
//
// Die zurückgegebene Funktion kann mit [setup.AndThen] mit einer weiteren Funktion kombiniert werden.
func theFixture(db *sql.DB) error {
	return model.ExecuteScripts(db, model.CreateScript, model.FixtureScript)
}

// func xFixture testx.SetupFunc {
// 	return testx.SetupFunc(model.CreateTable).AndThen(model.SetupFixture)
// }

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
	{
		DingRef: model.DingRef{
			Id:     1,
			Name:   "Paprika",
			Code:   "111",
			Anzahl: 1,
		},
		Beschreibung: "Eine Planzengattung, die zur Familie der Nachtschattengewächse gehört",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 18:48:01")),
	},

	{
		DingRef: model.DingRef{
			Id:     2,
			Name:   "Gurke",
			Code:   "222",
			Anzahl: 2,
		},
		Beschreibung: "",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:05:02")),
	},

	{
		DingRef: model.DingRef{
			Id:     3,
			Name:   "Tomate",
			Code:   "333",
			Anzahl: 3,
		},
		Beschreibung: "",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:06:03")),
	},
}
