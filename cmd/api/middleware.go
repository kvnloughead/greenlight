package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()

		// If no limiter exists for current IP, add it to the map of clients.
		if _, ok := clients[ip]; !ok {
			limiter := rate.NewLimiter(2, 4)
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

		next.ServeHTTP(w, r)
	})
}
