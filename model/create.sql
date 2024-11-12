CREATE TABLE dinge(
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  code VARCHAR(100) NOT NULL,
  name TEXT NOT NULL,
  anzahl INTEGER NOT NULL,
  aktualisiert DATETIME NOT NULL
);

CREATE INDEX idx_dinge_aktualisiert ON dinge(aktualisiert)
