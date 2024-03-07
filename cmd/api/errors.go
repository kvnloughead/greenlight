package main

import (
	"fmt"
	"net/http"
)

// logError logs an error message, as well as the request method and URL.
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI() // returns /path?query from the request URL
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// errorResponse sends arbitrary, JSON formatted errors to the client.
// It accepts a status code and a message of any type, wrapping the message in
// a JSON object with key "error". The result is sent using app.writeJSON.
//
// If app.writeJSON encounters an error, the function logs the error and sends
// a blank response with a 500 status code.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverErrorResponse logs an unexpected error at runtime.
// It logs the detailed error message, and uses app.errorResponse to send a 500
// Internal Server Error with a generic error message to the client.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "the server encountered a problem and couldn't process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

// notFoundResponse sends JSON response with a 404 status code.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "the requested resource cannot be found"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

// methodNotAllowedResponse sends a JSON response with a 405 status code.
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("the %s method is not allowed for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

// badRequestResponse sends a JSON response with a 400 status code. It accepts
// an error argument and includes it in the response.
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse sends a JSON response with a 422 status code. It
// accepts a map of errors and their messages and sends them in the response.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// editConflictResponse sends a JSON response with a 409 status code and a
// message that indicates a conflict while attempting to edit a resource.
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	msg := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, msg)
}

// rateLimitExceededReponse sends a JSON response with a 429 status code and a
// message that indicates that the rate limit has been exceeded.
func (app *application) rateLimitExceededReponse(w http.ResponseWriter, r *http.Request) {
	msg := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, msg)
}

// notFoundResponse sends JSON response with a 404 status code.
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// The invalidAuthenicationTokenResponse helper sends JSON response with a 401
// status code and "invalid authentication token" message. It also sets the
// "WWW-Authenticate" header to "Bearer" to remind the client that a bearer
// token is expected for authentication.
//
// This response helper is intended for use as a generic response to invalid
// attempts at authentication. For example, if the authentication header was
// missing or malformed.
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	msg := "invalid authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// An authenticationRequiredResponse is sent with a 401 status code when an
// unauthenticated user attempts to access a resource that requires
// authentication.
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	msg := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// An activationRequiredResponse is sent with a 403 status code when an
// unactivated user attempts to access a resource that requires activation.
func (app *application) activationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	msg := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}

// An permissionRequiredResponse is sent with a 403 status code when a user
// attempts to access a resource that they don't have permission to access.
func (app *application) permissionRequiredResponse(w http.ResponseWriter, r *http.Request) {
	msg := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}
