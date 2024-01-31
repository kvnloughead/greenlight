package main

import (
	"fmt"
	"net/http"
)

// Logs an error message, as well as the request method and URL.
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI() // returns /path?query from the request URL
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// Helper function for sending arbitrary, JSON formatted errors to the client. Accepts a status code and a message of any type. Wraps the message in a JSON object with the key "error", and sends the result with app.writeJSON.
//
// If an error occurs, it is logged, and a 500 status code is sent.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Helper function used when an unexpected error occurs at runtime. It logs the detailed error message, and uses the errorResponse helper to send a 500 Internal Server Error with a generic error message to the client.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {
	app.logError(r, err)

	msg := "the server encountered a problem and couldn't process your request"
	app.errorResponse(w, r, status, msg)
}

// Sends a 404 Not Found status code and JSON response to the client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "the requested resource cannot be found"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

// Sends a 405 Method Not Allowed error and a JSON response to the client.
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("the %s method is not allowed for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}
