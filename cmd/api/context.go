package main

import (
	"context"
	"net/http"

	"github.com/kvnloughead/greenlight/internal/data"
)

// The contextKey type is a custom string type for request context keys.
type contextKey string

var userContextKey = contextKey("user")

// The contextSetUser method accepts a request and a user struct as arguments,
// adds the user to the request's context with a key of "user", and returns a
// copy of the request.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// The contextGetUser method retrieves the value of the request context's user
// field, converts it to a *data.User, and returns it.
//
// This function should only be called when a User struct value is expected
// to be in the request context. If one is not found, there is a panic.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
