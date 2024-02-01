package main

import (
	"net/http"
	"time"

	"github.com/kvnloughead/greenlight/internal/data"
)

// createMovie handles POST requests to the /v1/movies endpoint.
func (app *application) createMovie(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Movie created\n"))
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
