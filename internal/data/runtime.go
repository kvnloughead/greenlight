package data

import (
	"fmt"
	"strconv"
)

// Runtime is a type that represents a movie's runtime in minutes.
// It satisfies the json.Marshaler interface, marshalling the runtime into
// JSON as to "<runtime> mins".
type Runtime int32

// MarshalJSON marshals an int32 runtime into JSON as "<runtime> mins".
// It returns the string as bytes, including the double quotes.
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	jsonValue = strconv.Quote(jsonValue)
	return []byte(jsonValue), nil
}
