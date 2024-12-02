CREATE TABLE dinge(
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  code VARCHAR(100) NOT NULL,
  name TEXT NOT NULL,
  anzahl INTEGER NOT NULL,
  beschreibung TEXT,
  aktualisiert DATETIME NOT NULL
);

CREATE INDEX idx_dinge_aktualisiert ON dinge(aktualisiert);
CREATE UNIQUE INDEX idx_dinge_code ON dinge(code);

CREATE VIRTUAL TABLE fulltext USING fts5(
  code,
  name,
  beschreibung);

CREATE TABLE photos(
  photo BLOB NOT NULL,
  mime_type VARCHAR(100) NOT NULL,
  dinge_id INTEGER NOT NULL REFERENCES dinge
);

CREATE UNIQUE INDEX idx_photos_dinge_id ON photos(dinge_id);
