package model_test

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"image/color"
	"image/draw"
	"iter"
	"reflect"
	"slices"
	"testing"
	"time"

	"image"

	"github.com/haschi/dinge/model"
	"github.com/haschi/dinge/sqlx"
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

func TestRepository_GetPhotoById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name       string
		args       args
		iterations int
		want       int
		wantErr    bool
	}{
		{
			name:       "get single photo",
			args:       args{id: 1, ctx: context.Background()},
			iterations: 1,
			want:       0,
			wantErr:    false,
		},
		{
			name:       "get multiple photos",
			args:       args{id: 1, ctx: context.Background()},
			iterations: 100,
			want:       0,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, theFixture, func(t *testing.T, db *sql.DB) {

				time.Sleep(100 * time.Millisecond)
				errorChan := make(chan error)

				t.Log("Number of iterations:", tt.iterations)

				r, err := model.NewRepository(db, system.RealClock{})

				if err != nil {
					t.Fatal(err)
					return
				}
				type iterationKey string
				for i := 0; i < tt.iterations; i++ {
					id := int64(i)
					go func() {
						// Damit der TransactionManager richtig funktioniert, benötigt jeder
						// Aufruf einen eigenen Context
						ctx := context.WithValue(context.Background(), iterationKey("iteration"), i)
						_, err = r.GetPhotoById(ctx, (id%3)+1)
						errorChan <- err
					}()
				}

				for i := 0; i < tt.iterations; i++ {
					err := <-errorChan
					if err != nil {
						t.Error(err)
					}
				}

				if r.Tm.Count() != 0 {
					t.Errorf("TransactionManager.Count() = %v; want %v", r.Tm.Count(), 0)
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
					Id:       dinge[0].Id,
					Code:     dinge[0].Code,
					Name:     dinge[0].Name,
					Anzahl:   dinge[0].Anzahl + 42,
					PhotoUrl: dinge[0].PhotoUrl,
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

func TestRepository_PhotoAktualisieren(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 600, 400))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	type fields struct {
		Clock model.Clock
	}
	type args struct {
		ctx context.Context
		id  int64
		im  image.Image
	}
	tests := []struct {
		name        string
		precodition func(*model.Repository) int64
		fields      fields
		args        args
		wantErr     error
	}{
		{
			name: "insert photo",
			args: args{ctx: context.Background(), im: img},
			precodition: func(r *model.Repository) int64 {
				res := must(r.Insert(context.Background(), "444", 1))
				if !res.Created {
					t.Fatal("Neues Ding hätte erzeugt werden müssen")
				}
				return res.Id
			},
		},
		{
			name:    "update photo",
			args:    args{ctx: context.Background(), id: 1, im: img},
			wantErr: nil,
		},
		{
			name:    "foreign key violation",
			args:    args{ctx: context.Background(), id: 4, im: img},
			wantErr: model.ErrNoRecord,
		},
		{
			name:    "bad image data",
			args:    args{ctx: context.Background(), id: 1, im: nil},
			wantErr: model.ErrInvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withDatabase(t, theFixture, func(t *testing.T, db *sql.DB) {
				r, err := model.NewRepository(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
				}

				if tt.precodition != nil {
					tt.args.id = tt.precodition(r)
				}

				err = r.PhotoAktualisieren(tt.args.ctx, tt.args.id, tt.args.im)
				if !errors.Is(err, tt.wantErr) {

					t.Errorf("Repository.PhotoAktualisieren() error = %v, want %v", err, tt.wantErr)
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
					Id:       dinge[0].Id,
					Name:     "Pepperoni",
					Code:     dinge[0].Code,
					Anzahl:   dinge[0].Anzahl,
					PhotoUrl: dinge[0].PhotoUrl,
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
					Id:       dinge[0].Id,
					Name:     "Pepperoni",
					Code:     dinge[0].Code,
					Anzahl:   dinge[0].Anzahl,
					PhotoUrl: dinge[0].PhotoUrl,
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

	type args struct {
		limit int
		query string
		sort  string
	}

	tests := []struct {
		name         string
		fields       fields
		precondition testx.SetupFunc
		args         args
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
			args:         args{limit: 0},
			want:         []model.DingRef{},
		},
		{
			name:         "limit is less than the number of stored items",
			fields:       model.NewRepository,
			precondition: theFixture,
			args:         args{limit: 2},
			want:         mapToDingeRef([]model.Ding{dinge[2], dinge[1]}),
		},
		{
			name:         "limit is greater than the number of stored items",
			fields:       model.NewRepository,
			precondition: theFixture,
			args:         args{limit: 4},
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

			args: args{limit: 4},
			want: mapToDingeRef([]model.Ding{
				{
					DingRef: model.DingRef{
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
			withDatabase(t, tt.precondition, func(t *testing.T, db *sql.DB) {
				r, err := tt.fields(db, system.RealClock{})
				if err != nil {
					t.Fatal(err)
				}

				got, err := r.GetLatest(context.Background(), tt.args.limit, tt.args.query, tt.args.sort)
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

// Don't use cache=shared.
// See https://github.com/mattn/go-sqlite3/issues/1179#issuecomment-1638083995
var dataSource = sqlx.ConnectionString("test.db",
	sqlx.MODE_MEMORY,
	sqlx.CACHE_Shared,
	sqlx.JOURNAL_WAL,
	sqlx.FK_ENABLED,
)

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

// withDatabase stellt eine Ausführungsumgebung bereit, in der eine Testfunktion mit Datenbank ausgeführt werden kann.
func withDatabase(t *testing.T, setupFn testx.SetupFunc, testFn testFunc) {
	t.Helper()
	t.Log("open database", dataSource)
	db, err := sql.Open("sqlite3", dataSource)

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
var dinge = []model.Ding{
	{
		DingRef: model.DingRef{
			Id:       1,
			Name:     "Paprika",
			Code:     "111",
			Anzahl:   1,
			PhotoUrl: "/photos/1",
		},
		Beschreibung: "Eine Planzengattung, die zur Familie der Nachtschattengewächse gehört",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 18:48:01")),
	},

	{
		DingRef: model.DingRef{
			Id:       2,
			Name:     "Gurke",
			Code:     "222",
			Anzahl:   2,
			PhotoUrl: "/photos/2",
		},
		Beschreibung: "",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:05:02")),
	},

	{
		DingRef: model.DingRef{
			Id:       3,
			Name:     "Tomate",
			Code:     "333",
			Anzahl:   3,
			PhotoUrl: "/photos/3",
		},
		Beschreibung: "",
		Aktualisiert: must(time.Parse(time.DateTime, "2024-11-13 19:06:03")),
	},
}
