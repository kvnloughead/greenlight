package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

/*
Set up router with httprouter. Available routes.

	```
	GET     /v1/healthcheck   Show application information
	GET     /v1/movies        Show details of all movies
	POST    /v1/movies        Create a new movie
	GET     /v1/movies/:id    Show details of a specific movie
	PUT     /v1/movies/:id    Update details of a specific movie
	DELETE  /v1/movies/:id    Delete a specific movie
	```
*/
func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheck)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovie)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovie)

	return router
}
