package data

import (
	"database/sql"
	"errors"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
	"github.com/lib/pq"
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

// Insert adds a new record to the movie table. It accepts a pointer to a
// Movie struct and runs an INSERT query. The id, created_at, and version fields
// are generated automatically.
func (m MovieModel) Insert(movie *Movie) error {
	// The query returns the system-generated id, created_at, and version fields
	// so that we can assign them to the movie struct argument.
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// The args slice contains the fields provided in the movie struct arguement.
	// Note that we are converting the string slice movie.Genres to an array the
	// is compatible with the genres field's text[] type.
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// QueryRow executes the query, passing the fields from args as placeholders.
	// The system-generated values are then scanned into the movie struct.
	return m.DB.QueryRow(query, args...).Scan(
		&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get retrieves a a specific record in the movies table by its ID. If the ID
// argument is less then 1, or if there is no movie with a matching ID in the
// database, and ErrRecordNotFound is returned. If a movie is found, a pointer
// to the corresponding Movie struct is returned.
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies WHERE ID = $1`

	var movie Movie

	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Update updates a specific record in the movies table. All fields must be
// provided by the caller, not only the fields to be changed. The record's
// version field is incremented by 1.
//
// Returns a sql.ErrNoRows error if no mathing record was found.
func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
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
