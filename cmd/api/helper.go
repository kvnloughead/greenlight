package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	// Parse ID from request params as int64.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("ID must be a positive integer")
	}

	return id, nil
}

// Marshals the data argument into JSON, then prepares and sends the response. The status code is always applied, and the Content-type header is set to application/json. Additional headers can optionally be specified.
//
// Errors are simply returned to the caller.
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
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
