INSERT INTO dinge(name, code, anzahl, beschreibung, aktualisiert)
VALUES ('Paprika', '111', 1,
  'Eine Planzengattung, die zur Familie der Nachtschattengewächse gehört',
  '2024-11-13 18:48:01'),
  ('Gurke', '222', 2, '', '2024-11-13 19:05:02'),
  ('Tomate', '333', 3, NULL, '2024-11-13 19:06:03');

INSERT INTO FULLTEXT (rowid, code, name, beschreibung)
SELECT id, code, name, beschreibung
FROM dinge;

INSERT INTO photos(photo, mime_type, dinge_id)
VALUES (X'0123456789', 'image/png', 1),
  (X'1234567890', 'image/jpeg', 2),
  (X'2345678901', 'image/webp', 3);

INSERT INTO history(operation, 'count', created, dinge_id)
VALUES (1, 1, '2024-11-13 18:48:01', 1),
  (1, 2, '2024-11-13 19:05:02', 2),
  (1, 3, '2024-11-13 19:06:03', 3);
