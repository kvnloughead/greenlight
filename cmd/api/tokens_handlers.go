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
