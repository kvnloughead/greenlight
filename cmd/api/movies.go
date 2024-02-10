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
// Request bodies are validated by ValidateMovie. A failedValidattionResponse
// error is sent if one or more fields fails validation.
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

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	data.ValidateMovie(v, movie)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Specify the API location of the created resource.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"data": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
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
		app.serverErrorResponse(w, r, err)
		return
	}
}
