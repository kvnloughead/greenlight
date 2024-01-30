package main

import (
	"net/http"
	"time"

	"github.com/kvnloughead/greenlight/internal/data"
)

// Handler fork GET /v1/healthcheck.
// Responds with info about the application, including version and the environment it is running in.
func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server couldn't process your request.", http.StatusInternalServerError)
		return
	}
}

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
		http.NotFound(w, r)
		app.logger.Error("handlers: ID must be a positive integer")
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
		app.logger.Error(err.Error())
		http.Error(w, "The server couldn't process your request.", http.StatusInternalServerError)
		return
	}
}
