package validation

type FormData[T any] struct {
	ValidationErrors ErrorMap
	Form             T
}
