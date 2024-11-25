package model

import "time"

// Clock ist eine Schnittstelle, die das aktuelle Datum und die aktuelle Uhrzeit bereitstellt.
type Clock interface {
	Now() time.Time
}
