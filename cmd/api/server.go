package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// serve creates and configures an instance of http.Server and returns the
// result of calling its ListenAndServe method.
func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	app.logger.Info(
		"Starting server",
		"port",
		app.config.port,
		"env",
		app.config.env,
	)

	return srv.ListenAndServe()
}
