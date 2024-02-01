package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// envelope is a type used for wrapping JSON responses to ensure a consistent
// response structure. It is a map with string keys and values of any type.
//
// It is commonly in handlers and middleware to wrap responses. For example,
// error responses are typically wrapped like this:
//
//	envelope{"error": "detailed error message"}
type envelope map[string]any

// readIdParam reads an ID param from the request context and parses it as an
// int64. If the ID doesn't parse to a positive integer, an error is returned.
func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("ID must be a positive integer")
	}

	return id, nil
}

// writeJSON marshals the data into JSON, then prepares and sends the response.
// The response is sent with
//
//  1. The "Content-type: application/json" header.
//  2. The status code that was supplied as an argument.
//
// Errors are simply returned to the caller.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Marshal data map into JSON for the response, indenting for readability.
	js, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	// Loop through the headers map and add each one to the ResponseWriter's header map. If headers is nil, this loop will simply be skipped.
	for k, v := range headers {
		w.Header()[k] = v
	}

	// Add Content-type header and status code. Then write JSON to response, appending a newline for QOL.
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	w.Write(append(js, '\n'))

	return nil
}
