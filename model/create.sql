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
