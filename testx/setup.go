package testx

import "database/sql"

// SetupFunc ist eine Function zum Herstellen der Vorbedingung für einen Testfall mit Datenbank
//
// SetupFunc findet in tabellengetriebenen Tests Verwendung, bei denen für jeden Testfall eine neue Datenbank mit testspezifischen Inhalten benötigt wird. Die Testfälle können ohne Datenbank zusammengestellt werden, in dem mit SetupFunc Funktionen kombiniert werden und erst ausgeführt werden, wenn der Testfall bereits in Ausführung ist und die Datenbank angelegt wurde.
type SetupFunc func(*sql.DB) error

// AndThen kombiniert die Funktion setup mit der Funktion then zu einer neuen Funktion.
//
// Wenn die resultierende Funktion aufgerufen wird, dann wird die then Funktion nur dann aufgerufen, wenn die setup Funktion erfolgreich war.
func (fn SetupFunc) AndThen(then SetupFunc) SetupFunc {
	return func(d *sql.DB) error {
		if err := fn(d); err != nil {
			return err
		}
		return then(d)
	}
}
