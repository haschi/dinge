INSERT INTO dinge(name, code, anzahl, beschreibung, aktualisiert)
VALUES
  ('Paprika', '111', 1, 'Eine Planzengattung, die zur Familie der Nachtschattengewächse gehört', '2024-11-13 18:48:01'),
  ('Gurke', '222', 2, '', '2024-11-13 19:05:02'),
  ('Tomate', '333', 3, NULL, '2024-11-13 19:06:03')
  -- ('Möhre', '444', 4, 'Eine Pflanzenart in der Familie der Doldenblütler', '2024-11-29 20:33:04')
  ;

INSERT INTO fulltext (rowid, code, name, beschreibung)
SELECT id, code, name, beschreibung from dinge;
