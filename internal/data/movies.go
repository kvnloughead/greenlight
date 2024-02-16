package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
	"github.com/lib/pq"
)

// Duration to use for SQL operation timeouts.
const queryTimeout = 3 * time.Second

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

// createTimeoutContext accepts a time duration and returns a context and cancel
// function with a timeout of that duration.
//
// The caller should defer calling the cancel() function.
func createTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return ctx, cancel
}

// GetAll retrieves a slice of movies from the database. The slice can be
// filtered and sorted via several optional query parameters.
//
//   - title: if provided, fuzzy matches on the movie's title.
//   - genres: if provided, only movies that have each of the provided genres
//     are included.
//   - sort: the key to sort by. Prepend with '-' for descending order. Defaults
//     to ID, ascending.
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	query := fmt.Sprintf(`
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('english', title)
					 @@ plainto_tsquery('english', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := createTimeoutContext(queryTimeout)
	defer cancel()

	// Retrieve matching rows from database.
	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres))
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Defer closing after handling errors.

	movies := []*Movie{}

	// Iterate through rows, reading each record in an entry in a Movie slice.
	for rows.Next() {
		var m Movie
		err = rows.Scan(
			&m.ID,
			&m.CreatedAt,
			&m.Title,
			&m.Year,
			&m.Runtime,
			pq.Array(&m.Genres),
			&m.Version,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &m)
	}

	// rows.Err() will contain any errors that occurred during iteration.
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return movies, nil
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

	ctx, cancel := createTimeoutContext(queryTimeout)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(
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

	ctx, cancel := createTimeoutContext(queryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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

// Update updates a specific record in the movies table. The caller should
// check for the existence of the record to be updated before calling Update.
// The record's version field is incremented by 1 after update.
//
// Prevents edit conflicts by verifying that the version of the record in the
// UPDATE query is the same as the version of the movie argument. In case of
// an edit conflict, an ErrEditConflict error is returned.
func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := createTimeoutContext(queryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		// An sql.ErrNoRows is returned if there are no matching records. Since we
		// know that the record exists already, this can be assumed to be due to a
		// version mismatch (hence an edit conflict).
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete deletes a specific record from the movies table. Returns an
// ErrNoRecordFound error if no record is found.
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM movies WHERE id = $1`

	ctx, cancel := createTimeoutContext(queryTimeout)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows are effected, then there was no record found.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

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
