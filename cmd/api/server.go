package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// serve creates and configures an instance of http.Server and returns the
// result of calling its ListenAndServe method.
//
// serve also establishes a coroutine that listens for SIGTERM and SIGINT
// signals. If either are found, the server's Shutdown() method is invoked,
// which gracefully shuts down the server.
func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutDownErr := make(chan error)

	go func() {
		// quit is a channel that carries values of type os.Signal. signal.Notify()
		// listens for SIGINT and SIGTERM signals, relaying them to the quit channel
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from channel quit. Blocks until a signal is received.
		s := <-quit

		// Log a message and quit application if SIGINT or SIGTERM is caught.
		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server, passing any errors to shutDownErr channel.
		shutDownErr <- srv.Shutdown(ctx)
	}()

	app.logger.Info(
		"Starting server",
		"port",
		app.config.port,
		"env",
		app.config.env,
	)

	// If an http.ErrServerClosed is returned by ListenAndServe() we ignore it
	// here, as it indicates a graceful shutdown has begun.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Check for problems with the graceful shutdown of our application.
	err = <-shutDownErr
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
