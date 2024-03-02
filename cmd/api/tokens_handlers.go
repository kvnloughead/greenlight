package main

import (
	"errors"
	"net/http"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
	"github.com/kvnloughead/greenlight/internal/data"
)

// The createActivationToken function handles POST request to the
// /v1/tokens/activation endpoint. It expects a JSON request body containing
// an email field. The following error responses are sent.
//
//   - badRequestResponse, if the response body can't be read
//   - failedValidationResponse, if the email isn't valid, or if the user is
//     already activated
//   - notFoundResponse, if there is no user with that email
//   - serverErrorResponse, for all other errors
//
// If the request is successful, a new activation token is created and added to
// the tokens table, an background process is spawned to send the user a
// confirmation email, and an http.StatusAccepted response is sent.
func (app *application) createActivationToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	if !v.Valid() {
		v.AddError("email", "no matching email found")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if user.Activated {
		v.AddError("email", "user already activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 72*time.Hour, data.Activation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := struct{ Token *data.Token }{Token: token}

		err = app.mailer.Send(user.Email, "token_activation.tmpl", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	env := envelope{"message": "an email will be sent to you containing activation instructions"}

	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// The createAuthenticationToken function handles POST requests to the
// /v1/tokens/authentication endpoint. It generates stateful authentication
// tokens.
//
// The user's credentials (email and password) are read from the body of the
// request and validated. If a user with the given email address isn't found,
// or if the password is incorrect, a 401 response is sent by the
// app.invalidCredentials helper.
//
// If the credentials check out we generate a token with a 24 hour expiry and
// an "authentication" scope. This token is then sent to the client in a JSON
// response with the following format:
//
//	{
//	    "authentication_token": {
//	        "token": "N4AN76GAQIXFKRIVRRKW463X5Q",
//	        "expiry": "2024-03-03T17:12:34.711714248-05:00"
//	    }
//	}
func (app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	// Read user credentials from request body into the input struct.
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the request body's fields.
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve user from users table. If a record is not found, we send a 401
	// "invalid credentials" response.
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Check if the password matches the hash. If not, we send a 401 "invalid
	// credentials" response.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// If the credentials check out we generate a token with a 24 hour expiry and
	// an "authentication" scope.
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour,
		data.Authentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w,
		http.StatusCreated,
		envelope{"authentication_token": token},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
