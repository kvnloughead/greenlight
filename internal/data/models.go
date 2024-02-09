package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound is an error returned by Get, Delete, and Update methods
	// if there is no corresponding record.
	ErrRecordNotFound = errors.New("record not found")
)

// Models is a struct that wraps all of our models.
type Models struct {
	Movies MovieModel
}

// NewModels returns an empty instance of our Model struct.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
