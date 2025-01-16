CREATE TABLE photos(
  photo BLOB NOT NULL,
  mime_type VARCHAR(100) NOT NULL,
  dinge_id INTEGER NOT NULL REFERENCES dinge
);

CREATE UNIQUE INDEX idx_photos_dinge_id ON photos(dinge_id);
