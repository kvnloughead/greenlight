package data

import (
	"fmt"
	"strconv"
)

// Movie runtime in minutes. Satisfies the json.Marshaler interface and marshals into JSON as to "<runtime> mins".
type Runtime int32

// Marshals runtime into JSON as "<runtime> mins". Returns this string as bytes, including the double quotes. Without the quotes, the string would be misinterprented.
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	jsonValue = strconv.Quote(jsonValue)
	return []byte(jsonValue), nil
}
