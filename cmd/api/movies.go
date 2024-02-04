package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kvnloughead/greenlight/internal/data"
)

// createMovie handles POST requests to the /v1/movies endpoint.
func (app *application) createMovie(w http.ResponseWriter, r *http.Request) {
	// Struct to store the data from the responses body. The struct's fields must
	// be exported to use it with json.NewDecoder.
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// Decode responses body into the struct above. Keys will be mapped to struct
	// fields with matching json tags, but if that fails an attempt to match the
	// struct keys (case-insensitively). Unmatched keys in the decoded data will
	// be ignored. Unmatched keys in the decode destination will be given their
	// zero value.
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Dump decoded body as an HTTP response.
	fmt.Fprintf(w, "%+v\n", input)
}

// showMovie handles GET requests to the /v1/movies/:id endpoint.
func (app *application) showMovie(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Spartacus",
		Runtime:   90,
		Genres:    []string{"drama", "war"},
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
