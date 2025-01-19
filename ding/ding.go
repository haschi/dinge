package ding

import "time"

type Ding struct {
	DingRef
	Beschreibung string
	Allgemein    string
	Aktualisiert time.Time
}

// DingRef repräsentiert ein Ding in der Übersicht.
type DingRef struct {
	Id       int64
	Name     string
	Code     string
	Anzahl   int
	PhotoUrl string
}

func (d DingRef) Equal(other DingRef) bool {
	return d.Id == other.Id &&
		d.Name == other.Name &&
		d.Code == other.Code &&
		d.Anzahl == other.Anzahl &&
		d.PhotoUrl == other.PhotoUrl
}
