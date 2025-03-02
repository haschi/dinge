![workflow](https://github.com/haschi/dinge/actions/workflows/go.yml/badge.svg)

# dinge

Dinge verwalten

## Datenbank anlegen

Bevor du die Anwendung benutzen kannst, musst du eine Datenbank anlegen. Gebe dazu folgenden Befehl in der Konsole ein:

```bash
sqlite3 dinge.db < model/create.sql
```

Anschließend sollte sich die Datei *dinge.db* im aktuellen Verzeichnis befinden.

## Anwendung starten

Die Anwendung startest du mit dem Befehl

```bash
go run .
```

oder du erzeugst eine ausführbare Datei mit dem Befehl

```code
go build .
```

Dies erzeugt die Datei *dinge*. Du startest die Anwendung dann mit dem Befehl

```bash
./dinge
```

Mit `./dinge --help` erhälst du eine übersicht der möglichen Parameter.

## Anwendung beenden

Gibst du die Tastenkombination <key>Strg</key>-<key>C</key> in der Konsole ein, beendest du damit die Anwendung.

## Development

### Literatur

[SQLTools Erweiterung für Visual Studio Code](https://vscode-sqltools.mteixeira.dev/en/home)
