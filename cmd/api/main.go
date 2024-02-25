package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/kvnloughead/greenlight/internal/data"
	"github.com/kvnloughead/greenlight/internal/mailer"
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

	// limiter is a struct containing configuration for our rate limiter.
	limiter struct {
		rps     float64 // Requests per second. Defaults to 2.
		burst   int     // Max request in burst. Defaults to 4.
		enabled bool    // Defaults to true.
	}

	// smtp is a struct containing configuration for our SMTP server.
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// The application struct is used for dependency injection.
type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer

	// The WaitGroup instance allows us to track goroutines in progress, to
	// prevent shutdown until they are all completed. No need for initialization,
	// the zero-valued sync.WaitGroup is useable, with counter set to 0.
	wg sync.WaitGroup
}

func main() {
	// Parse CLI flags into config struct (to be added to dependencies).
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "The API's HTTP port.")
	flag.StringVar(&cfg.env,
		"env",
		"development",
		"Environment (development|staging|production)")

	// Read DB-related settings from CLI flags.
	flag.StringVar(&cfg.db.dsn,
		"db-dsn",
		os.Getenv("GREENLIGHT_DB_DSN"),
		"Postgresql DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql max connection idle time")

	// Read rate-limter-related settings from CLI flags.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second per IP")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter max requests in a burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read SMTP related settings from CLI flags. The defaults are derived from
	// the Mailtrap server we are using for testing.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP server port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "d2d67cf14feb94", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "62eabaae7885b8", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@github.com/kvnloughead/greenlight>", "SMTP sender")

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
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username,
			cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
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
