package validation

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ErrorMap = map[string]string

type Form struct {
	Request          *http.Request
	ValidationErrors ErrorMap
}

func NewForm(req *http.Request) *Form {
	return &Form{
		Request:          req,
		ValidationErrors: map[string]string{},
	}
}

func (f *Form) Close() {
	if f != nil {
		// TODO: Handle error
		f.Request.Body.Close()
	}
}

type ValidationError struct {
	key     string
	message string
}

type FieldType interface{ int | int64 | string | bool }

func Field[T FieldType](key string, extractor ExtractorFunc[T], validators ...ValidationFunc[T]) ScanFunc {
	return func(s url.Values) *ValidationError {
		rawValue := s.Get(key)
		value, err := extractor(key, rawValue)

		if err != nil {
			return err
		}

		for _, validator := range validators {
			validator(value)
		}
		return nil
	}
}

type ScanFunc func(s url.Values) *ValidationError

type ValidationFunc[T FieldType] func(value T) *ValidationError

type ExtractorFunc[T FieldType] func(key string, value string) (T, *ValidationError)

func (f *Form) Scan(scanners ...ScanFunc) error {

	if err := f.Request.ParseForm(); err != nil {
		return err
	}

	for _, scanner := range scanners {
		if err := scanner(f.Request.Form); err != nil {
			f.ValidationErrors[err.key] = err.message
		}
	}

	return nil
}

func (f Form) IsValid() bool {
	return len(f.ValidationErrors) == 0
}

/* ScanFunc */

func String(reference *string) ExtractorFunc[string] {
	return func(key, value string) (string, *ValidationError) {
		*reference = value
		return value, nil
	}
}

func Int64(reference *int64) ExtractorFunc[int64] {
	return func(key, value string) (int64, *ValidationError) {
		result, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, &ValidationError{key: key, message: "Kein Zahl"}
		}
		*reference = result
		return result, nil
	}
}

func Integer(reference *int) ExtractorFunc[int] {

	return func(key, value string) (int, *ValidationError) {
		result, err := strconv.Atoi(value)
		if err != nil {
			return 0, &ValidationError{key: key, message: "Keine Zahl"}
		}
		*reference = result
		return result, nil
	}
}

func IsNotBlank(value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			message: "Das Feld ist leer, darf es aber nicht sein",
		}
	}

	return nil
}

func Min(lowerbound int) ValidationFunc[int] {
	return func(value int) *ValidationError {
		if value < lowerbound {
			return &ValidationError{
				message: "Wert darf nicht kleiner sein als x",
			}
		}
		return nil
	}
}
