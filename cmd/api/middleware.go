package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
	"github.com/kvnloughead/greenlight/internal/data"
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
	// Struct client contains data corresponding to a client IP. It has a rate
	// limiter property, and a lastSeen property used to remove unused clients
	// from the clients map.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Start background goroutine to remove old entries from the clients map.
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			// If no limiter exists for current IP, add it to the map of clients.
			if _, ok := clients[ip]; !ok {
				limiter := rate.NewLimiter(
					rate.Limit(app.config.limiter.rps),
					app.config.limiter.burst,
				)
				clients[ip] = &client{limiter: limiter}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededReponse(w, r)
				return
			}

			// We can't defer unlocking this mutext, because it wouldn't occur until all
			// downstream handlers have retured.
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

// The authenticate middleware authenticates a user based on the token provided
// in the authorization header. The header should be of the form "Bearer
// <token>". The token should be 26 bytes long.
//
// 401 Unauthorized responses are sent if the authorization header is
// malformed, if the token is invalid, or if a user record corresponding to the
// token isn't found.
//
// If everything checks out, the user's data is added to the request context.
// Otherwise, the anonymous user is added to the request context.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The "Vary: Authorization" header indicates to caches that the response
		// may vary based on the value of the request's Authorization header.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			// If there is no authorization header, add anonymous user to the context.
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Split the header and return a 401 if not in the format "Bearer <token>".
		parts := strings.Split(authorizationHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := parts[1]

		// Validate that the token is 26 bytes long.
		v := validator.New()
		data.ValidateTokenPlaintext(v, token)
		if !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get user from DB. If record isn't found we send a 401 response.
		user, err := app.models.Users.GetForToken(data.Authentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Add user to request context and call the next handler.
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}
