package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// The routes function initializes and returns an http.Handler with all the
// route definitions for the application. It uses httprouter for routing
// requests to their corresponding handlers based on the HTTP method and path.
//
// The defined routes are as follows:
//
//   - GET    /v1/healthcheck   				 Show application information.
//
//   - GET    /v1/movies								 Show details of a subset of	 movies.
//
//   - POST   /v1/movies								 Create a new movie.
//
//   - GET    /v1/movies/:id	  				 Show details of a specific movie.
//
//   - PATCH  /v1/movies/:id						 Update details of a specific movie.
//
//   - DELETE /v1/movies/:id	  				 Delete a specific movie.
//
//   - POST   /v1/users         				 Register a new user.
//
//   - PUT    /v1/users/activated     	 Activates a user.
//
//   - POST   /v1/tokens/activation   	 Generate a new activation token.
//
//   - POST   /v1/tokens/authentication  Generate an authentication token.
//
// This function also sets up custom error handling for scenarios where no
// route is matched (404 Not Found) and when a method is not allowed for a
// given route (405 Method Not Allowed), using the custom error handlers
// defined in api/errors.go.
//
// Finally, the router is wrapped with the recoverPanic middleware to handle any
// panics that occur during request processing.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Set custom error handlers for 404 and 405 errors.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheck)

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requireActivatedUser(app.listMovies))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requireActivatedUser(app.createMovie))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requireActivatedUser(app.showMovie))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requireActivatedUser(app.updateMovie))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requireActivatedUser(app.deleteMovie))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUser)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUser)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation",
		app.createActivationToken)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication",
		app.createAuthenticationToken)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
