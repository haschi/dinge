package model_test

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
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

type fakeClock struct {
	ding model.Ding
}

func (c fakeClock) Now() time.Time {
	return c.ding.Aktualisiert
}

func setupFixture(db *sql.DB, dinge []model.Ding) error {

	for _, ding := range dinge {

		r := model.Repository{DB: db, Clock: fakeClock{ding: ding}}

		result, err := r.Insert(context.Background(), ding.Code, ding.Anzahl)
		if err != nil {
			return err
		}

		if !result.Created {
			return errors.New("ding bereits vorhanden")
		}

		if result.Id != ding.Id {
			return errors.New("id missmatch")
		}

		if err = r.NamenAktualisieren(context.Background(), result.Id, ding.Name); err != nil {
			return err
		}
	}

	// Nur zu Testzwecken

	rows, err := db.Query("SELECT id, name, code, anzahl, aktualisiert FROM dinge")
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var ding model.Ding
		err = rows.Scan(&ding.Id, &ding.Name, &ding.Code, &ding.Anzahl, &ding.Aktualisiert)
		if err != nil {
			return err
		}

		fmt.Println(ding)
	}

	return nil
}

type setup func(*sql.DB) error

type testFunc func(*testing.T, *sql.DB)

func thefixture(dinge []model.Ding) setup {
	return func(d *sql.DB) error {
		return errors.Join(
			initializeDatabase(d),
			setupFixture(d, dinge))
	}
}

func withDatabase(t *testing.T, setup setup, testfn testFunc) {
	t.Helper()

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		t.Fatal("can not open database", err)
	}

	defer db.Close()

	if setup == nil {
		t.Fatal("no setup function provided")
	}

	if err := setup(db); err != nil {
		t.Fatal("can not setup fixture", err)
	}

	testfn(t, db)
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

var dinge = []model.Ding{
	{Id: 1, Name: "Paprika", Code: "111", Anzahl: 1, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 18:48:01"))},

	{Id: 2, Name: "Gurke", Code: "222", Anzahl: 2, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:05:02"))},

	{Id: 3, Name: "Tomate", Code: "333", Anzahl: 3, Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:06:06"))},
}

func TestRepository_GetById(t *testing.T) {

	withDatabase(t, thefixture(dinge), func(t *testing.T, db *sql.DB) {

		type fields struct {
			DB *sql.DB
		}
		type args struct {
			id int64
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    model.Ding
			wantErr bool
		}{
			{
				name:    "empty database",
				fields:  fields{DB: db},
				args:    args{id: 0},
				want:    model.Ding{},
				wantErr: true,
			},
			{
				name:    "ding 111",
				fields:  fields{DB: db},
				args:    args{id: 1},
				want:    dinge[0],
				wantErr: false,
			},
			{
				name:    "No database",
				fields:  fields{DB: nil},
				args:    args{id: 1},
				want:    model.Ding{},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := model.Repository{
					DB:    tt.fields.DB,
					Clock: model.RealClock{},
				}
				got, err := r.GetById(tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.GetById() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Repository.GetById() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestRepository_GetByCode(t *testing.T) {

	withDatabase(t, thefixture(dinge), func(t *testing.T, db *sql.DB) {
		type fields struct {
			DB *sql.DB
		}
		type args struct {
			code string
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    model.Ding
			wantErr bool
		}{
			{
				name:    "Get existing ding",
				fields:  fields{DB: db},
				args:    args{code: dinge[0].Code},
				want:    dinge[0],
				wantErr: false,
			},
			{
				name:    "Get unknown ding",
				fields:  fields{DB: db},
				args:    args{code: "unknown"},
				want:    model.Ding{},
				wantErr: true,
			},
			{
				name:    "Get without database",
				fields:  fields{DB: nil},
				args:    args{code: "doesn't matter"},
				want:    model.Ding{},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := model.Repository{
					DB: tt.fields.DB,
				}
				got, err := r.GetByCode(tt.args.code)
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

type FixedClock struct {
	Timestamp time.Time
}

func (c FixedClock) Now() time.Time {
	return c.Timestamp
}

type repositoryProvider func(*sql.DB, model.Clock) model.Repository

func TestRepository_MengeAktualisieren(t *testing.T) {

	setDb := func(db *sql.DB, clock model.Clock) model.Repository { return model.Repository{DB: db, Clock: clock} }

	type args struct {
		ctx   context.Context
		code  string
		menge int
	}
	tests := []struct {
		name    string
		fields  repositoryProvider
		args    args
		want    model.Ding
		wantErr bool
	}{
		{
			name:   "Update Paprika",
			fields: setDb,
			args:   args{code: dinge[0].Code, menge: 42, ctx: context.Background()},
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
			fields:  setDb,
			args:    args{code: "unknown", menge: 42, ctx: context.Background()},
			want:    model.Ding{},
			wantErr: true,
		},
		{
			name:    "Update too much",
			fields:  setDb,
			args:    args{code: dinge[0].Code, menge: -(dinge[0].Anzahl + 1), ctx: context.Background()},
			want:    model.Ding{},
			wantErr: true,
		},
		{
			name:    "no database",
			fields:  func(db *sql.DB, clock model.Clock) model.Repository { return model.Repository{} },
			args:    args{code: "doesn't matter", menge: 0, ctx: context.Background()},
			want:    model.Ding{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			withDatabase(t, thefixture(dinge), func(t *testing.T, db *sql.DB) {
				r := tt.fields(db, FixedClock{Timestamp: tt.want.Aktualisiert})
				id, err := r.MengeAktualisieren(tt.args.ctx, tt.args.code, tt.args.menge)

				if (err != nil) != tt.wantErr {
					t.Errorf("Repository.MengeAktualisieren() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if err != nil {
					return
				}

				got, err := r.GetById(id)
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

	withDatabase(t, thefixture(dinge), func(t *testing.T, db *sql.DB) {

		type fields struct {
			DB *sql.DB
		}
		type args struct {
			ctx    context.Context
			code   string
			anzahl int
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    model.InsertResult
			wantErr bool
		}{
			{
				name:    "without database",
				fields:  fields{DB: nil},
				args:    args{ctx: context.Background(), code: "doesn't matter", anzahl: 42},
				wantErr: true,
			},
			{
				name:   "Insert new ding",
				fields: fields{DB: db},
				args:   args{ctx: context.Background(), code: "QWERT", anzahl: 1},
				want:   model.InsertResult{Id: int64(len(dinge) + 1), Created: true},
			},
			{
				name:   "Insert existing ding",
				fields: fields{DB: db},
				args:   args{ctx: context.Background(), code: dinge[0].Code, anzahl: 1},
				want:   model.InsertResult{Id: dinge[0].Id, Created: false},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := model.Repository{
					DB:    tt.fields.DB,
					Clock: model.RealClock{},
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
		}
	})
}

// var fixture setup = func(db *sql.DB) error {
// 	return setupFixture(db, dinge)
// }

func TestRepository_NamenAktualisieren(t *testing.T) {
	withDatabase(t, thefixture(dinge), func(t *testing.T, db *sql.DB) {
		type fields struct {
			DB *sql.DB
		}
		type args struct {
			id        int64
			name      string
			timestamp time.Time
			ctx       context.Context
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    model.Ding
			wantErr bool
		}{
			{
				name:    "without database",
				fields:  fields{},
				args:    args{id: 0, name: "", timestamp: must(time.Parse(time.DateTime, "1970-01-01 00:00:00"))},
				wantErr: true,
			},
			{
				name:   "change Paprika",
				fields: fields{DB: db},
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
				name:    "update unknown",
				fields:  fields{DB: db},
				args:    args{ctx: context.Background(), id: 42, name: "doesn't matter"},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := model.Repository{
					DB:    tt.fields.DB,
					Clock: FixedClock{Timestamp: tt.args.timestamp},
				}
				if err := r.NamenAktualisieren(tt.args.ctx, tt.args.id, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("Repository.NamenAktualisieren() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					got, err := r.GetById(tt.args.id)
					if err != nil {
						t.Fatal("unexpected condition")
					}
					if !reflect.DeepEqual(got, tt.want) {
						t.Errorf("Repository.GetById() = %v, want %v", got, tt.want)
					}
				}
			})
		}
	})
}

func repositoryWithoutDatabase(_ *sql.DB) model.Repository {
	return model.Repository{}
}

func repositoryWithDatabase(db *sql.DB) model.Repository {
	return model.Repository{DB: db}
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
			precondition: func(d *sql.DB) error {
				if err := thefixture(dinge)(d); err != nil {
					return err
				}
				repository := model.Repository{DB: d, Clock: FixedClock{Timestamp: must(time.Parse(time.DateTime, "2024-11-13 19:58:05"))}}
				repository.MengeAktualisieren(context.Background(), dinge[1].Code, 1)
				return nil
			},
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
