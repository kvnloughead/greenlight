package data

import (
	"database/sql"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
)

// Movie is a struct representing data for a single movie entry.
type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

// MovieModel struct wraps an sql.DB connection pool and implements
// basic CRUD operations.
type MovieModel struct {
	DB *sql.DB
}

// Insert adds a movie to the database.
func (m MovieModel) Insert(movie *Movie) error {
	return nil
}

// Get retrieves a a specific record in the movies table by its ID.
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Update updates a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// Delete deletes a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {
	return nil
}

// ValidateMovie validates the fields of a Movie struct. The fields must meet
// the following requirements:
//
//   - Title, Year, Runtime, and Genres are required.
//   - Title must be less than 500 bytes.
//   - Year must be between 1888 and the present.
//   - Runtime must be a positive integer.
//   - There must be between 1 and 5 unique genres.
func ValidateMovie(v *validator.Validator, m *Movie) {

	v.Check(m.Title != "", "title", "must be provided")
	v.Check(len(m.Title) < 500, "title", "must be less than 500 bytes")

	v.Check(m.Year != 0, "year", "must be provided")
	v.Check(m.Year >= 1888, "year", "must be after 1888")
	v.Check(m.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(m.Runtime != 0, "runtime", "must be provided")
	v.Check(m.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(m.Genres != nil, "genres", "must be provided")
	v.Check(len(m.Genres) >= 1, "genres", "must be at least 1 genre")
	v.Check(len(m.Genres) <= 5, "genres", "must be no more than 5 genres")
	v.Check(validator.Unique(m.Genres), "genres", "must not contain duplicate values")
}
