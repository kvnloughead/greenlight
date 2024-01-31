package main

import (
	"net/http"
	"time"

	"github.com/kvnloughead/greenlight/internal/data"
)

// Handler for POST /v1/movies.
// Creates a new movie and adds it the the database.
func (app *application) createMovie(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Movie created\n"))
}

// Handler for GET /v1/movies/:id.
// Shows details for the movie with the specified ID.
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