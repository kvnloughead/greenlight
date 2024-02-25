package main

import (
	"errors"
	"fmt"
	"net/http"

	validator "github.com/kvnloughead/greenlight/internal"
	"github.com/kvnloughead/greenlight/internal/data"
)

// registerUser handles POST requests to the /v1/users endpoint. The request
// body is decoded by the app.readJSON helper. See that function for details
// about error handling.
//
// The request body must contain a name, email, and password. Request bodies
// are validated by data.ValidateUser. A failedValidationResponse error is sent
// if one or more fields fails validation, or if the email is a duplicate.
//
// A hash is generated from the plaintext password via bcrypt and stored in the
// DB.
//
// On successful registration, a welcome email is sent via app.mailer.
func (app *application) registerUser(w http.ResponseWriter, r *http.Request) {
	// Struct to store the data from the responses body. The struct's fields must
	// be exported to use it with json.NewDecoder.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy info from request body into a new user struct. Below, we set the
	// password by calling the user struct's Password.Set method.
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Validate user. Email uniqueness is checked on attempted insert.
	v := validator.New()
	data.ValidateUser(v, user)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert new record into DB, if possible.
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Lauch goroutine to send a welcome email.
	go func() {
		defer func() {
			// Catch panics and log resulting errors.
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.Error(err.Error())
		}
	}()

	// Write JSON response.
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
