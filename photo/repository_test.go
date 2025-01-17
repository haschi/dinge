package photo_test

import (
	"context"
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"testing"
	"time"

	"github.com/haschi/dinge/ding"
	"github.com/haschi/dinge/photo"
	"github.com/haschi/dinge/sqlx"
	"github.com/haschi/dinge/system"
)

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
			withTransactionManager(t, theFixture, func(t *testing.T, tm sqlx.TransactionManager) {

				time.Sleep(100 * time.Millisecond)
				errorChan := make(chan error)

				repository := photo.Repository{
					Clock: system.RealClock{},
					Tm:    tm,
				}

				type iterationKey string
				for i := 0; i < tt.iterations; i++ {
					id := int64(i)
					go func() {
						// Damit der TransactionManager richtig funktioniert, benötigt jeder
						// Aufruf einen eigenen Context
						ctx := context.WithValue(context.Background(), iterationKey("iteration"), i)
						_, err := repository.GetPhotoById(ctx, (id%3)+1)
						errorChan <- err
					}()
				}

				for i := 0; i < tt.iterations; i++ {
					err := <-errorChan
					if err != nil {
						t.Error(err)
					}
				}

				if repository.Tm.Count() != 0 {
					t.Errorf("TransactionManager.Count() = %v; want %v", repository.Tm.Count(), 0)
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
		Clock photo.Clock
	}
	type args struct {
		ctx context.Context
		id  int64
		im  image.Image
	}
	tests := []struct {
		name        string
		precodition func(sqlx.TransactionManager) int64
		fields      fields
		args        args
		wantErr     error
	}{
		{
			name: "insert photo",
			args: args{ctx: context.Background(), im: img},
			precodition: func(tm sqlx.TransactionManager) int64 {
				repo := &ding.Repository{
					Clock: system.RealClock{},
					Tm:    tm,
				}

				res := must(repo.Insert(context.Background(), "444", 1))
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
			wantErr: photo.ErrNoRecord,
		},
		{
			name:    "bad image data",
			args:    args{ctx: context.Background(), id: 1, im: nil},
			wantErr: photo.ErrInvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTransactionManager(t, theFixture, func(t *testing.T, tm sqlx.TransactionManager) {
				repository := &photo.Repository{
					Clock: system.RealClock{},
					Tm:    tm,
				}

				if tt.precodition != nil {
					tt.args.id = tt.precodition(tm)
				}

				err := repository.PhotoAktualisieren(tt.args.ctx, tt.args.id, tt.args.im)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Repository.PhotoAktualisieren() error = %v, want %v", err, tt.wantErr)
				}
			})
		})
	}
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

// must stoppt die Ausführung, wenn err nicht nil ist. Ansonsten liefert die Funktion den Wert value zurück.
//
// Die Funktion sollte nur in Testcode verwendet werden.
func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

// thefixture liefert eine Funktion, mit der die Vorbedingung eines Tests hergestellt werden kann
//
// Die zurückgegebene Funktion initialisiert die Datenbank und füllt diese mit Testdaten, wenn sie aufgerufen wird.
//
// Die zurückgegebene Funktion kann mit [setup.AndThen] mit einer weiteren Funktion kombiniert werden.
func theFixture(db *sql.DB) error {
	return sqlx.ExecuteScripts(db, ding.CreateScript, photo.CreateScript, ding.FixtureScript, photo.FixtureScript)
}
