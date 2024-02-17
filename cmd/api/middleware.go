package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
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

// rateLimit is a middleware that limits the number of requests to an average of
// 2 per second, with bursts of up to 4 seconds.
//
// If the limit is exceeded, a 429 Too Many Request response is sent to the
// client.
func (app *application) rateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(2, 4)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if !limiter.Allow() {
				app.rateLimitExceededReponse(w, r)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}
