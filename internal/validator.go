package validator

import (
	"regexp"
	"slices"
)

// EmailRX is a regex pattern matching a valid email, recommended by W3C.
// https://html.spec.whatwg.org/multipage/input.html#valid-e-mail-address
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator is a struct for validating JSON responses. It contains several
// validation methods and an Error map to store error messages.
type Validator struct {
	Errors map[string]string
}

// New returns a Validator instance with an empty Errors map.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Validator.Valid returns true if the validator's Errors map is empty.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Validator.AddError adds an error to the validator's Errors map (as long as
// if it doesn't already exist.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Validator.Check adds an error to Validator.Errors if ok is false.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Returns true if the string matches the regex.
func Matches(s string, rx *regexp.Regexp) bool {
	return rx.MatchString(s)
}

// Returns true if the value matches one of the permittedValues.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Unique accepts a slice of any comparable type and returns true if all
// values in it are unique.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, v := range values {
		uniqueValues[v] = true
	}

	return len(values) == len(uniqueValues)
}
