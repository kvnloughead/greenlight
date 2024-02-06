package main

import (
	"fmt"
	"net/http"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
	"github.com/kvnloughead/greenlight/internal/data"
)

// createMovie handles POST requests to the /v1/movies endpoint. The request
// body is decoded by the app.readJSON helper. See that function for details
// about error handling.
//
// Requests must meet the following validation.
//
//   - Title, Year, Runtime, and Genres are required.
//   - Title must be less than 500 bytes.
//   - Year must be between 1888 and the present.
//   - Runtime must be a positive integer.
//   - There must be between 1 and 5 unique genres.
func (app *application) createMovie(w http.ResponseWriter, r *http.Request) {
	// Struct to store the data from the responses body. The struct's fields must
	// be exported to use it with json.NewDecoder.
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.Title != "", "title", "must be provided")
	v.Check(len(input.Title) < 500, "title", "must be less than 500 bytes")

	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be after 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(input.Genres != nil, "genres", "must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must be at least 1 genre")
	v.Check(len(input.Genres) <= 5, "genres", "must be no more than 5 genres")
	v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

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
