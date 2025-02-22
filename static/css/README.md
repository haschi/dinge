# Cascading Style Sheets

Style Sheets für die Anwendung

## Anforderungen:

- CSS Reset / Defaults
- Wichtige Parameter als Variablen:
  - Schriftgröße
  - Farben
  - Schriftart
  - ...
- Typografie
  - Systemzeichensatz verwenden *system-ui*
  - Schriftgröße aus dem Desktop Thema übernehmen.
  - Schriftgöße für Überschriften bestimmen.
- Seitenleiste - Globale Steuerelemente
  - Kopfzeile und Navigation der Seitenleiste sind immer sichtbar.
  - Kopfzeile der Seitenleiste enthält
    - Links: Schnellaktionen
    - Mitte: Anwendungsname
    - Rechts: Anwendungsmenü
  - Navigation unterhalb der Kopfzeile
    - Untereinander in der Seitenleiste
  - Anwendungsmenü
    - Auf der rechten Seite in der Kopfzeile des Seitenleiste
    - Mit Button aktivieren
    - Menü als Popover anzeigen
    - Schließen durch
      - Auswahl eines Elements innerhalb des Popovers
      - Erneutes klicken des Aktivierungsbuttons
      - Klicken in einen Bereich außerhalb des Popovers
    - Systemfarben
  - Aktionen
    - Auf der linken Seite in der Kopfzeile des Seitenleiste
    - 0 bis n Aktionen möglich
    - Buttons oder Linkbuttons.
    - Systemfarben
- Kopfzeile des Inhaltsbereichs
  - Mitte: Titel der Seite = Name des Navigationslinks der Seitenleiste
- Übersicht gleichartiger Elemente
  - Liste
    - `table` Element
  - Rasteransicht
    - `ul` Element
  - Umschalten zwischen den Modi
  - Suche / Filter
    - Mittleres Element der Kopfzeile des Inhaltsbereichs
  - Sortierung
    - Ansichtmenü
    - Popovermechanik
    - Rechte Seite der Kopfzeile des Inhaltsbereichs
    -
- Nur Text für:
  - Leeren Inhalt, keine Daten
  - Lizenz
  - Anleitung
  - Codebeispiele zum Beispiel für
    - Dokumentation der HTML / CSS UI-Muster
  - Überschrift
  - Paragraphen
  - Hervorhebungen
  - *Breakout* für Beispiele
- Datenansicht
  - Daten eines einzelnen Elements anzeigen
  - Struktur der Daten als Definitionsliste (`dl`)
- Formular
  - Daten eines einzelnen Elements eingeben / ändern.
- Vertikaler Flow
  - Konsequent den Abstand zum Vorgänger berücksichtigen
  - Nicht den Abstand zum Nachfolger
- Akkordeon
  - Zum Ausblenden von Details
  - Möglich in
    - Datenansicht
    - Formular
    - Nur Text
  - Systemfarben, Farbverlauf nutzen
  -

## Abgrenzung / Nicht zu implementieren

- Lightmode später implementieren
- Keine Skalierung der Schriftgöße und Abstände


## Micro Aufgabenliste

- [X] Eingabe-Element type=file: Gutes Markup finden, dann CSS.
- [ ] An Chromium anpassen
- [ ] Popups (in der Kopfzeile) stylen.
- [X] Abstände der Schaltflächen im Flow.

- [ ] Kopfzeile: Sucheingabe stylen
- [ ] Eingabefelder: Mehrzeilig, wenn das Label hochgeklappt wird und einzeilige, bei denen der Wert rechtsbündig bearbeitet wird, unterscheiden. Nicht anhand des Typs, sondern HTML Struktur.
- [X] Typographie: Zeilenabstände in Paragraphen vergrößern.
- [ ] Farbe für Gefahr. Berechnen aus der Akkzentfarbe. Gleiche Helligkeit, Sättigung, aber ein Rotton.
- [ ] Formular: Berechnung der Zeilenhöhe anpassen an Änderungen in Typographie.
- [X] Formular: Auswahl: Boxen um das Label ändern ihre Größe.
- [X] Abstände der Listenelemente in geordneten Listen (ol). Abstände sind derzeit zu klein.
- [ ] Gemeinsame Werte in Variablen auslagern. z.B. Hintergrundfarben, Hervorhebungsfarben.
- [ ] Abgeleitete Werte Berechnen statt Konstanten.

## Tabellen

- [X] margin-bottom in Tabellen eliminieren. Ich will nur margin-top wegen der Konsistenz.
- [X] Gaps in Tabellenzeilen entfernen. Gaps zwischen den Zeilen bleiben.
- [X] Größe von Bildern automatisch an Zeilenhöhe anpassen. Keine explizite Größenangabe. Der Browser soll das Bild automatisch skalieren.
- [X] Text sekundärer Spalten dunkler.
- [X] Tabellenzeile zur Navigation anklickbar.????
- [ ] Tabellenzeilen nur anklickbar, wenn Text des Verweises leer ist.

### Buttons und Links

- [ ] Klärung: Semantik und Markup von Buttons: Primär, Sekundär, Destruktiv: Innerhalb!
- [ ] Alle Buttons durchstylen
- [ ] Alle Links durchstylen

### Fokus

- [X] Fokussierte Links (normale Links) hervorheben
- [X] Fokussierte Zusammenfassung von Details mit der Akzentfarbe hervorheben. (details, summary)
- [X] Fokussierte Eingabefelder type=file sind noch ungestyled.
- [X] Seitenleiste: Fokusierte Elemente hervorheben und umrahmen
- [X] Eingabe-Element type=color: Wenn Fokus, hervorheben.
- [X] Fokussierung der Zeilen von Definitionslisten.
- [X] Fokussierte Karten gestalten.
- [X] input[type=file] darf nicht umrahmt werden, wenn es innerhalb eines label Elements erscheint, das umrahmt ist. Ebenso date, datetime-local, time, select, range, option, color, button.
- [X] Reguläre Links dürfen nicht mit der gleichen Farbe umrahmt werden, wie die Textfarbe.
- [X] Bei Eingabefelder, Detailansicht-Zeilen, die den Fokus haben und mit der Akzentfarbe umrahmt sind, sollen auch der Hintergrund hervorgehoben werden,
- [ ] Fokussierte Steuerelemente in den Kopfzeilen richtig stellen.

### Formular

- [X] textarea hinzufügen.
- [ ] Validierungsfehler ausgeben.
- [ ] Eingabefelder mit ungültigen Daten markieren. Pseudoklasse.

### Detailansicht

- [X] Hintergrundfarbe der Zeilen ist zu hell.
- [X] Trennlinie zwischen den Zeilen, ähnlich Formular erforderlich.
- [X] Definitionsliste (Detailansicht): Chevron anzeigen, wenn die jeweilige Zeile ein Navigationselement enthält, um den Wert zu bearbeiten oder weitere Details anzuzeigen.

### Akkordeon

- [ ] Hover Effekt für Summary. (Ggf. erst nach Auslagern der Farbwerte in Variablen. Im Moment herscht Chaos.)
- [X] Abstände, so dass alle Inhalte vernünftig aussehen.
