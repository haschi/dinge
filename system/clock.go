package system

import "time"

// RealClock liefert die Systemzeit
type RealClock struct{}

func (c RealClock) Now() time.Time {
	return time.Now()
}
