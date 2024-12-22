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
CREATE VIRTUAL TABLE fulltext USING fts5(code, name, beschreibung);
CREATE TABLE photos(
  photo BLOB NOT NULL,
  mime_type VARCHAR(100) NOT NULL,
  dinge_id INTEGER NOT NULL REFERENCES dinge
);
CREATE UNIQUE INDEX idx_photos_dinge_id ON photos(dinge_id);
CREATE TABLE history(
  operation INTEGER NOT NULL REFERENCES operation,
  count INTEGER NOT NULL,
  created DATETIME NOT NULL,
  dinge_id INTEGER NOT NULL REFERENCES dinge
);
CREATE INDEX idx_history_created ON history(created);
CREATE INDEX idx_history_dingeId ON history(dinge_id);
CREATE TABLE operation(
  id INTEGER NOT NULL PRIMARY KEY,
  name TEXT NOT NULL
);
INSERT INTO operation(id, name)
VALUES(1, 'new'),
  (2, "add"),
  (3, "delete");
