package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidRuntimeFormat is returned when an input runtime is not of the
// format '<int32> mins".
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Runtime is a type that represents a movie's runtime in minutes.
//
// It satisfies the json.Marshaler and json.Unmarshaler interfaces, marshalling
// to and from "<runtime> mins", where "<runtime>" is an int32.
type Runtime int32

// MarshalJSON marshals an int32 runtime into JSON as "<runtime> mins".
// It returns the string as bytes, including the double quotes.
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	jsonValue = strconv.Quote(jsonValue)
	return []byte(jsonValue), nil
}

// UnmarshalJSON converts a json value of the form "<int32> mins" into
// instances of the Runtime time type. an ErrInvalidRuntimeFormat is returned
// in these cases:
//
//  1. If the value isn't wrapped in double quotes.
//  2. If it is not of the format "<runtime> mins".
//  3. If "<runtime>" can't be converted into an int32.
func (r *Runtime) UnmarshalJSON(jsonVal []byte) error {

	s, err := strconv.Unquote(string(jsonVal))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Split(s, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	n, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(n)

	return nil
}
