package main

import (
	"fmt"
	"net/http"
)

// Defers a function to be called in all events. If a panic occurred, the "Connection: close" header is set to close the server, and an internal server error is sent to the client.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				// err has type any so must be converted to error
				app.serverErrorResponse(w, r, http.StatusInternalServerError,
					fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})

}
