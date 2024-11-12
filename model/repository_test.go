package model

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed create.sql
var create string

const dataSourceName = "file::memory:?cache=shared"

func initializeDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(create)
	if err != nil {
		t.Fatal("can not initialze database", err)
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		t.Fatal("can not query tables")
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			t.Fatal("can not scan row")
		}
		t.Logf("Table found: %v", name)
	}
}

func setupFixture(t *testing.T, db *sql.DB, dinge []Ding) {
	t.Helper()

	for _, ding := range dinge {
		r := Repository{DB: db}
		result, err := r.Insert(context.Background(), ding.Code, ding.Anzahl)
		if err != nil {
			t.Fatal("can not insert ding:", err)
		}

		if !result.Created {
			t.Fatal("ding bereits vorhanden")
		}

		if result.Id != ding.Id {
			t.Fatal("id mismatch")
		}

		if err = r.NamenAktualisieren(result.Id, ding.Name); err != nil {
			t.Fatal("can not update name", err)
		}
	}
}

func withDatabase(t *testing.T, dinge []Ding, testfn func(t *testing.T, db *sql.DB)) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		t.Fatal("can not open database", err)
	}

	defer db.Close()

	initializeDatabase(t, db)
	setupFixture(t, db, dinge)
	testfn(t, db)
}

var dinge = []Ding{
	{Id: 1, Name: "Paprika", Code: "111", Anzahl: 1},
	{Id: 2, Name: "Gurke", Code: "222", Anzahl: 2},
	{Id: 3, Name: "Tomate", Code: "333", Anzahl: 3},
}

func TestRepository_GetById(t *testing.T) {
	withDatabase(t, dinge, func(t *testing.T, db *sql.DB) {

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
			want    Ding
			wantErr bool
		}{
			{
				name:    "empty database",
				fields:  fields{DB: db},
				args:    args{id: 0},
				want:    Ding{},
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
				want:    Ding{},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Repository{
					DB: tt.fields.DB,
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

	withDatabase(t, dinge, func(t *testing.T, db *sql.DB) {
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
			want    Ding
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
				want:    Ding{},
				wantErr: true,
			},
			{
				name:    "Get without database",
				fields:  fields{DB: nil},
				args:    args{code: "doesn't matter"},
				want:    Ding{},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Repository{
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

func TestRepository_MengeAktualisieren(t *testing.T) {
	type fields func(db *sql.DB) Repository

	setDb := func(db *sql.DB) Repository { return Repository{DB: db} }

	type args struct {
		ctx   context.Context
		code  string
		menge int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Ding
		wantErr bool
	}{
		{
			name:   "Update Paprika",
			fields: setDb,
			args:   args{code: dinge[0].Code, menge: 42, ctx: context.Background()},
			want: Ding{
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
			want:    Ding{},
			wantErr: true,
		},
		{
			name:    "Update too much",
			fields:  setDb,
			args:    args{code: dinge[0].Code, menge: -(dinge[0].Anzahl + 1), ctx: context.Background()},
			want:    Ding{},
			wantErr: true,
		},
		{
			name:    "no database",
			fields:  func(db *sql.DB) Repository { return Repository{} },
			args:    args{code: "doesn't matter", menge: 0, ctx: context.Background()},
			want:    Ding{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			withDatabase(t, dinge, func(t *testing.T, db *sql.DB) {
				r := tt.fields(db)
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

	withDatabase(t, dinge, func(t *testing.T, db *sql.DB) {

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
			want    InsertResult
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
				want:   InsertResult{Id: int64(len(dinge) + 1), Created: true},
			},
			{
				name:   "Insert existing ding",
				fields: fields{DB: db},
				args:   args{ctx: context.Background(), code: dinge[0].Code, anzahl: 1},
				want:   InsertResult{Id: dinge[0].Id, Created: false},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := Repository{
					DB: tt.fields.DB,
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

func TestRepository_NamenAktualisieren(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		id   int64
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Repository{
				DB: tt.fields.DB,
			}
			if err := r.NamenAktualisieren(tt.args.id, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Repository.NamenAktualisieren() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		id     int64
		anzahl int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Repository{
				DB: tt.fields.DB,
			}
			if err := r.Update(tt.args.id, tt.args.anzahl); (err != nil) != tt.wantErr {
				t.Errorf("Repository.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_GetLatest(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	tests := []struct {
		name    string
		fields  fields
		want    []Ding
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Repository{
				DB: tt.fields.DB,
			}
			got, err := r.GetLatest()
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetLatest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}
