package main

import (
	"fmt"
	"net/http"
)

// recoverPanic is a middleware that catches all panics in a handler chain.
// When a panic is caught, it is handled by
//  1. Setting the "Connection: close" header, to instruct go to shut down the
//     server after sending the response.
//  2. Sending a 500 Internal Server Error response containing the error from
//     the recovered panic.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				// err has type any so must be converted to error
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})

}
