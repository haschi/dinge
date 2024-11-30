package validation

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

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

type ScanFunc func(s url.Values) *ValidationError

type ErrorMap = map[string]string

type ValidationError struct {
	key     string
	message string
}

func (err ValidationError) Error() string {
	return err.message
}

func String(key string, value *string, validators ...ValidationFunc[string]) ScanFunc {
	return func(s url.Values) *ValidationError {
		*value = s.Get(key)

		return validate(key, *value, validators)
	}
}

func Integer(key string, value *int, validators ...ValidationFunc[int]) ScanFunc {
	return func(s url.Values) *ValidationError {
		i, err := strconv.Atoi(s.Get(key))
		if err != nil {
			return &ValidationError{key: key, message: "Keine Zahl"}
		}
		*value = i
		return validate(key, i, validators)
	}
}

func Integer64(key string, value *int64, validators ...ValidationFunc[int64]) ScanFunc {
	return func(s url.Values) *ValidationError {
		i, err := strconv.ParseInt(s.Get(key), 10, 64)
		if err != nil {
			return &ValidationError{key: key, message: "Keine Zahl"}
		}
		*value = i

		return validate(key, i, validators)
	}
}

func validate[T FieldType](key string, value T, validators []ValidationFunc[T]) *ValidationError {
	for _, validator := range validators {
		if err := validator(value); err != nil {
			return &ValidationError{key: key, message: err.Error()}
		}
	}

	return nil
}

type ValidationFunc[T FieldType] func(value T) error

type FieldType interface {
	int | int64 | string | bool
}

/* ScanFunc */

func IsNotBlank(value string) error {
	if strings.TrimSpace(string(value)) == "" {
		return errors.New("Das Feld ist leer, darf es aber nicht sein")
	}

	return nil
}

func MaxLength(max int) ValidationFunc[string] {
	return func(value string) error {
		if len(value) > max {
			return errors.New("Zuviele Zeichen für dieses Feld eingegeben")
		}
		return nil
	}
}

func StringOptions(options ...string) ValidationFunc[string] {
	return func(value string) error {
		if slices.Contains(options, value) {
			return nil
		}
		return errors.New("Ungültige Auswahl")
	}
}
func Min(lowerbound int) ValidationFunc[int] {
	return func(value int) error {
		if value < lowerbound {
			return errors.New("Wert darf nicht kleiner sein als x")
		}
		return nil
	}
}
