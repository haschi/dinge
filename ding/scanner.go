package ding

import (
	"github.com/haschi/dinge/webx"
)

type ScannerFormData struct {
	Title            string
	ActionUrl        string
	SubmitButtonText string

	Code    string
	Anzahl  int
	History []Event
}

func NewScannerFormData(code string, anzahl int, history []Event) webx.TemplateData[ScannerFormData] {
	return webx.TemplateData[ScannerFormData]{
		Scripts: []string{
			"/static/barcode.js",
		},
		Styles: []string{"/static/css/barcode.css"},
		FormValues: ScannerFormData{
			Title:            "Einlagern",
			ActionUrl:        "/dinge/",
			SubmitButtonText: "Einlagern",
			Code:             "",
			Anzahl:           1,
			History:          history,
		},
	}
}

func NewDestroyFormData(code string, anzahl int, history []Event) webx.TemplateData[ScannerFormData] {
	return webx.TemplateData[ScannerFormData]{
		Scripts: []string{
			"/static/barcode.js",
		},
		Styles: []string{"/static/css/barcode.css"},
		FormValues: ScannerFormData{
			Title:            "Entnehmen",
			ActionUrl:        "/dinge/delete",
			SubmitButtonText: "Entnehmen",
			Code:             code,
			Anzahl:           anzahl,
			History:          history,
		},
	}
}
