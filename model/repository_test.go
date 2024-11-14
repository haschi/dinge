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

	_ "github.com/mattn/go-sqlite3"
)

//go:embed create.sql
var create string

const dataSourceName = "file::memory:?cache=shared"

func initializeDatabase(db *sql.DB) error {

	_, err := db.Exec(create)
	if err != nil {
		return err
	}

	// rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	// if err != nil {
	// 	return err
	// }

	// defer rows.Close()

	// for rows.Next() {
	// 	var name string
	// 	if err = rows.Scan(&name); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func setupFixture(ctx context.Context, dinge []model.Ding) setup {

	return func(db *sql.DB) error {
		for _, ding := range dinge {

			r := model.Repository{DB: db, Clock: FixedClock{Timestamp: ding.Aktualisiert}}

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

// setup ist eine Function zum Herstellen der Vorbedingung für einen Testfall
type setup func(*sql.DB) error

// AndThen kombiniert die Funktion setup mit der Funktion then zu einer neuen Funktion.
//
// Wenn die resultierende Funktion aufgerufen wird, dann wird die then Funktion nur dann aufgerufen, wenn die setup Funktion erfolgreich war.
func (fn setup) AndThen(then setup) setup {
	return func(d *sql.DB) error {
		if err := fn(d); err != nil {
			return nil
		}
		return then(d)
	}
}

// testFunc deklariert eine Testfunktion, die eine Datenbank benutzt
type testFunc func(*testing.T, *sql.DB)

// thefixture liefert eine Funktion, mit der die Vorbedingung eines Tests hergestellt werden kann
//
// Die zurückgegebene Funktion initialisiert die Datenbank und füllt diese mit Testdaten, wenn sie aufgerufen wird.
//
// Die zurückgegebene Funktion kann mit [setup.AndThen] mit einer weiteren Funktion kombiniert werden.
func thefixture(dinge []model.Ding) setup {
	return setup(initializeDatabase).AndThen(setupFixture(context.Background(), dinge))
}

// withDatabase stellt eine Ausführungsumgebung bereit, in der eine Testfunktion ausgeführt werden kann, die eine Datenbank verwendet.
func withDatabase(t *testing.T, setup setup, testfn testFunc) {
	t.Helper()

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		t.Fatal("can not open database", err)
	}

	defer db.Close()

	if setup != nil {
		if err := setup(db); err != nil {
			t.Fatal("can not setup fixture", err)
		}
	}

	if testfn == nil {
		t.Fatal("no test function provided")
	}

	testfn(t, db)
}

var dinge = []model.Ding{
	{Id: 1, Name: "Paprika", Code: "111", Anzahl: 1, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 18:48:01"))},

	{Id: 2, Name: "Gurke", Code: "222", Anzahl: 2, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:05:02"))},

	{Id: 3, Name: "Tomate", Code: "333", Anzahl: 3, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:06:06"))},
}

func TestRepository_GetById(t *testing.T) {

	type args struct {
		id  int64
		ctx context.Context
	}
	tests := []struct {
		name         string
		fields       repositoryProvider
		precondition setup
		args         args
		want         model.Ding
		wantErr      bool
	}{
		{
			name:         "empty database",
			fields:       newRespository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background()},
			wantErr:      true,
		},
		{
			name:         "lese ding 111",
			fields:       newRespository,
			precondition: thefixture(dinge),
			args:         args{id: dinge[0].Id, ctx: context.Background()},
			want:         dinge[0],
			wantErr:      false,
		},
		{
			name:    "No database",
			fields:  newRespositoryWithoutDatabase,
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
		{
			name:         "closed database",
			fields:       newRespository,
			precondition: closeDatabase,
			wantErr:      true,
		},
		{
			name:    "context done",
			fields:  newRespository,
			args:    args{ctx: newCanceledContext()},
			wantErr: true,
		},
		{
			name:    "context nil",
			fields:  newRespository,
			args:    args{ctx: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r := tt.fields(db, model.RealClock{})
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
				name:    "without database",
				fields:  newRespositoryWithoutDatabase,
				wantErr: true,
			},
			{
				name:    "Get existing ding",
				fields:  newRespository,
				args:    args{ctx: context.Background(), code: dinge[0].Code},
				want:    dinge[0],
				wantErr: false,
			},
			{
				name:    "Get unknown ding",
				fields:  newRespository,
				args:    args{ctx: context.Background(), code: "unknown"},
				wantErr: true,
			},
			{
				name:    "without context",
				fields:  newRespository,
				args:    args{},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := tt.fields(db, model.RealClock{})
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
		setup   setup
		args    args
		want    model.Ding
		wantErr bool
	}{
		{
			name:   "Update Paprika",
			fields: newRespository,
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
			fields:  newRespository,
			setup:   thefixture(dinge),
			args:    args{code: "unknown", menge: 42, ctx: context.Background()},
			wantErr: true,
		},
		{
			name:    "Update too much",
			fields:  newRespository,
			setup:   thefixture(dinge),
			args:    args{code: dinge[0].Code, menge: -(dinge[0].Anzahl + 1), ctx: context.Background()},
			wantErr: true,
		},
		{
			name:    "no database",
			setup:   thefixture(dinge),
			fields:  newRespositoryWithoutDatabase,
			args:    args{code: "doesn't matter", menge: 0, ctx: context.Background()},
			wantErr: true,
		},
		{
			name:    "without context",
			fields:  newRespository,
			args:    args{code: dinge[0].Code, menge: 1, ctx: nil},
			wantErr: true,
		},
		{
			name:    "with canceled context",
			fields:  newRespository,
			setup:   thefixture(dinge).AndThen(closeDatabase),
			args:    args{code: "", menge: 0, ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			withDatabase(t, tt.setup, func(t *testing.T, db *sql.DB) {
				r := tt.fields(db, FixedClock{Timestamp: tt.want.Aktualisiert})
				id, err := r.MengeAktualisieren(tt.args.ctx, tt.args.code, tt.args.menge)

				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.MengeAktualisieren() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if err != nil {
					return
				}

				got, err := r.GetById(tt.args.ctx, id)
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

// func closeDatabase() setup {
// 	return func(d *sql.DB) error {
// 		return d.Close()
// 	}
// }

func closeDatabase(db *sql.DB) error {
	return db.Close()
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
		precondition setup
		args         args
		want         model.InsertResult
		wantErr      bool
	}{
		{
			name:         "without database",
			fields:       newRespositoryWithoutDatabase,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background(), code: "doesn't matter", anzahl: 42},
			wantErr:      true,
		},
		{
			name:         "closed database",
			fields:       newRespository,
			precondition: thefixture(dinge).AndThen(closeDatabase),
			args:         args{ctx: context.Background(), code: "doesn't matter", anzahl: 42},
			wantErr:      true,
		},
		{
			name:         "Insert new ding",
			fields:       newRespository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background(), code: "QWERT", anzahl: 1},
			want:         model.InsertResult{Id: int64(len(dinge) + 1), Created: true},
		},
		{
			name:         "Insert existing ding",
			fields:       newRespository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background(), code: dinge[0].Code, anzahl: 1},
			want:         model.InsertResult{Id: dinge[0].Id, Created: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r := tt.fields(db, model.RealClock{})
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
		precondition setup
		args         args
		want         model.Ding
		wantErr      bool
	}{
		{
			name:    "without database",
			fields:  newRespositoryWithoutDatabase,
			wantErr: true,
		},
		{
			name:         "closed database",
			precondition: closeDatabase,
			fields:       newRespository,
			args:         args{ctx: context.Background()},
			wantErr:      true,
		},
		{
			name:    "without context",
			fields:  newRespository,
			args:    args{ctx: nil},
			wantErr: true,
		},
		{
			name:         "change Paprika",
			fields:       newRespository,
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
			fields:       newRespository,
			precondition: thefixture(dinge),
			args:         args{ctx: context.Background(), id: 42, name: "doesn't matter"},
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r := tt.fields(db, FixedClock{Timestamp: tt.args.timestamp})
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

	type fields func(*sql.DB) model.Repository
	tests := []struct {
		name         string
		fields       fields
		precondition setup
		arg          int
		want         []model.Ding
		wantErr      bool
	}{
		{
			name:         "without database",
			fields:       repositoryWithoutDatabase,
			precondition: func(db *sql.DB) error { return nil },
			wantErr:      true,
		},
		{
			name:         "empty database",
			fields:       repositoryWithDatabase,
			precondition: thefixture([]model.Ding{}),
			want:         []model.Ding{},
		},
		{
			name:         "database closed",
			fields:       repositoryWithDatabase,
			precondition: closeDatabase,
			wantErr:      true,
		},
		{
			name:         "limit 0",
			fields:       repositoryWithDatabase,
			precondition: thefixture(dinge),
			arg:          0,
			want:         []model.Ding{},
		},
		{
			name:         "limit is less than the number of stored items",
			fields:       repositoryWithDatabase,
			precondition: thefixture(dinge),
			arg:          2,
			want:         []model.Ding{dinge[2], dinge[1]},
		},
		{
			name:         "limit is greater than the number of stored items",
			fields:       repositoryWithDatabase,
			precondition: thefixture(dinge),
			arg:          4,
			want:         []model.Ding{dinge[2], dinge[1], dinge[0]},
		},
		{
			name:   "items are sorted by date of change",
			fields: repositoryWithDatabase,
			precondition: thefixture(dinge).AndThen(func(d *sql.DB) error {
				repository := model.Repository{
					DB:    d,
					Clock: FixedClock{Timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:58:05"))},
				}
				_, err := repository.MengeAktualisieren(context.Background(), dinge[1].Code, 1)
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
				r := tt.fields(db)
				got, err := r.GetLatest(tt.arg)
				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.GetLatest() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
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
type repositoryProvider func(*sql.DB, model.Clock) model.Repository

// newRepository liefert ein neues [model.Repository] mit Datenbank und clock.
func newRespository(db *sql.DB, clock model.Clock) model.Repository {
	return model.Repository{
		DB:    db,
		Clock: clock,
	}
}

// newRepositoryWithoutDatabase liefert ein neues [model.Repository] ohne Datenbank.
func newRespositoryWithoutDatabase(db *sql.DB, clock model.Clock) model.Repository {
	return model.Repository{
		DB:    nil,
		Clock: clock,
	}
}

// must stoppt die Ausführung, wenn err wahr ist. Ansonsten liefert die Funktion den Wert value zurück.
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

func repositoryWithoutDatabase(_ *sql.DB) model.Repository {
	return model.Repository{}
}

func repositoryWithDatabase(db *sql.DB) model.Repository {
	return model.Repository{DB: db}
}
