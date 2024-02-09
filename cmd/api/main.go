package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kvnloughead/greenlight/internal/data"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// config is a struct containing configuration settings. These settings are
// specified as CLI flags when application starts, and have defaults provided
// in case they are omitted.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

// application is a struct used for dependency injection.
type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	// Parse CLI flags into config struct (to be added to dependencies).
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "The API's HTTP port.")
	flag.StringVar(&cfg.env,
		"env",
		"development",
		"Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn,
		"db-dsn",
		os.Getenv("GREENLIGHT_DB_DSN"),
		"Postgresql DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql max connection idle time")
	flag.Parse()

	// Create structured logger (to be added to dependencies).
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Open database connection.
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	app := application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("Starting server", "port", cfg.port, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB creates an sql.DB connection pool for the supplied DSN and returns it.
// If a connection can't be established within 5 seconds, an error is returned.
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	// Create a context with an empty parent context and a 5s timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to connect to the database within the 5s lifetime of the context.
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
