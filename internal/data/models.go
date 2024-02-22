package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound is an error returned by Get, Delete, and Update methods
	// if there is no corresponding record.
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is an error returned if there is a conflict when updating
	// a resource. It indicates that the resource was already changed or deleted
	// since the current request was initiated.
	ErrEditConflict = errors.New("edit conflict")
)

// Models is a struct that wraps all of our models.
type Models struct {
	Movies MovieModel
	Users  UserModel
}

// NewModels returns an empty instance of our Model struct.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
