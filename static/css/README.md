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
- [ ] Eingabe-Element type=color: Wenn Fokus, hervorheben.
- [ ] An Chromium anpassen
- [ ] Popups (in der Kopfzeile) stylen.
- [X] Abstände der Schaltflächen im Flow.
- [ ] Definitionsliste (Detailansicht): Chevron anzeigen, wenn die jeweilige Zeile ein Navigationselement enthält, um den Wert zu bearbeiten oder weitere Details anzuzeigen.
- [X] Fokussierung der Zeilen von Definitionslisten.
- [ ] Kopfzeile: Sucheingabe stylen
- [ ] Eingabefelder: Mehrzeilig, wenn das Label hochgeklappt wird und einzeilige, bei denen der Wert rechtsbündig bearbeitet wird, unterscheiden. Nicht anhand des Typs, sondern HTML Struktur.
- [X] Typographie: Zeilenabstände in Paragraphen vergrößern.
- [ ] Klärung: Semantik und Markup von Buttons: Primär, Sekundär, Destruktiv: Innerhalb!
- [ ] Farbe für Gefahr. Berechnen aus der Akkzentfarbe. Gleiche Helligkeit, Sättigung, aber ein Rotton.
- [ ] Formular: Berechnung der Zeilenhöhe anpassen an Änderungen in Typographie.
- [X] Formular: Auswahl: Boxen um das Label ändern ihre Größe.
- [X] Fokussierte Eingabefelder type=file sind noch ungestyled.
- [X] Fokussierte Zusammenfassung von Details mit der Akzentfarbe hervorheben. (details, summary)
- [ ] Seitenleiste: Fokusierte Elemente hervorheben und umrahmen
- [ ] Fokussierte Links (normale Links) hervorheben
- [ ] Abstände der Listenelemente in geordneten Listen (ol). Abstände sind derzeit zu klein.
- [ ] margin-bottom in Tabellen eliminieren. Ich will nur margin-top wegen der Konsistenz.
- [ ] Alle Buttons durchstylen.
