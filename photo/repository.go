package photo

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/png"

	"github.com/haschi/dinge/sqlx"
	"github.com/mattn/go-sqlite3"
)

type Repository struct {
	Clock Clock
	Tm    sqlx.TransactionManager
}

func (r Repository) GetPhotoById(ctx context.Context, dingId int64) ([]byte, error) {

	suchen := `SELECT photo FROM photos WHERE dinge_id = :id`

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var photo []byte
	row := tx.QueryRowContext(suchen, sql.Named("id", dingId))
	if err := row.Scan(&photo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return photo, err
	}
	return photo, tx.Commit()
}

// GetUrl liefert die URL des Photos eines Dings.
//
// [id] ist die id eines Dings.
func (r Repository) GetUrl(ctx context.Context, dingId int64) (string, error) {

	url := "/static/placeholder.svg"

	if ctx == nil {
		return url, errors.New("no context provided")
	}

	suchen := `
	SELECT COUNT(rowid)
	FROM photos
	WHERE dinge_id = :dingId
	`

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return url, err
	}

	defer tx.Rollback()

	row := tx.QueryRowContext(suchen, sql.Named("dingId", dingId))
	var count int
	if err := row.Scan(&count); err != nil {
		return url, err
	}

	if count == 1 {
		url = fmt.Sprintf("/photos/%v", dingId)
	}

	return url, tx.Commit()
}

func (r Repository) PhotoAktualisieren(ctx context.Context, id int64, image image.Image) error {

	if image == nil {
		return ErrInvalidParameter
	}

	// TODO in einen Thumbnail Service auslagern
	zuschnitt := Crop(image)
	thumbnail := Resize(zuschnitt)

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, thumbnail); err != nil {
		return err
	}

	timestamp := r.Clock.Now()

	statement := `
	INSERT INTO photos(photo, mime_type, dinge_id)
	VALUES(:photo, :mime_type, :id)
	ON CONFLICT (dinge_id)
	DO UPDATE SET photo = :photo, mime_type = :mime_type;
	`

	tx, err := r.Tm.BeginTx(ctx)
	if err != nil {
		return nil
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(statement,
		sql.Named("id", id),
		sql.Named("photo", buffer.Bytes()),
		sql.Named("mime_type", "image/png"),
		sql.Named("aktualisiert", timestamp))

	if err != nil {
		if e, ok := err.(sqlite3.Error); ok {
			if e.Code == sqlite3.ErrConstraint {
				return ErrNoRecord
			}
		}
		return err
	}

	return tx.Commit()
}

var ErrNoRecord = errors.New("no record found")
var ErrInvalidParameter = errors.New("invalid paramater")
