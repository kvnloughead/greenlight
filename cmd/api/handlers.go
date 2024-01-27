package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Handler fork GET /v1/healthcheck.
// Responds with info about the application, including version and the environment it is running in.
func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	// Marshal data map into JSON for the response.
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server can't process your request.", http.StatusInternalServerError)
		return
	}

	// Specify that the response is JSON and send it, appending a newline for QOL.
	w.Header().Set("Content-type", "application/json")
	w.Write(append(js, '\n'))
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
	fmt.Fprintf(w, "Showing movie number %d\n", id)
}
