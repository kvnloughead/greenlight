package data

import (
	"database/sql"
	"errors"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrDupKeyConstraintMsg stores the error message returned by postgres when
	// duplicate email constraint is violated.
	ErrDupKeyConstraintMsg = `pq: duplicate key value violates unique constraint "users_email_key"`

	// ErrDuplicateEmail is our custom duplicate email error.
	ErrDuplicateEmail = errors.New("duplicate email")
)

// User is a struct representing data for an individual user. The Password and
// Version fields are omitted from the JSON representation.
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int32     `json:"-"`
}

type UserModel struct {
	DB *sql.DB
}

// Insert adds a new record to the users table. It accepts a pointer to a
// User struct and runs an INSERT query. The id, created_at, and version fields
// are generated automatically.
//
// If a user already exists with the given email, an ErrDuplicateEmail error is
// returned.
func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID, &user.CreatedAt, &user.Version,
	)
	if err != nil {
		switch {
		case err.Error() == ErrDupKeyConstraintMsg:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetByEmail retrieves a user record with matching email.
//
// If no such record exists, it returns an ErrRecordNotFound error.
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
		where email = $1`

	// An empty struct to store the document returned by the query.
	var user User

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	// Run the query and populate the empty user struct. Since we have a unique
	// constraint on the email field, at most one row will be returned. If
	// multiple rows were found, the first row would be used.
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// The Update function updates an existing user document.
//
// The document's version is checked to eliminate edit conflicts. In these cases
// an ErrEditConflict is returned.
//
// In case of an attempt to change the email to an existing email address, an
// ErrDuplicateEmail is returned.
func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email $2, password_hash $3, 
			  activated $4, version = version + 1
		WHERE id = $5 and version = $6
		RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		// Since the document is know to exist, an ErrNoRows must be due to version
		// difference, indicating an edit conflict.
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		case err.Error() == ErrDupKeyConstraintMsg:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

// The password struct stores a password's plaintext representation and the
// computed hash. The plaintext field is a pointer to a string, so we can
// distinguish between non-existent passwords and empty string passwords.
type password struct {
	plaintext *string
	hash      []byte
}

// Set calculates a hash from the supplied plaintext password using bcrypt,
// storing the plaintext and hash in the fields of the calling password struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches determines whether the supplied password matches the calling struct's
// hash field, using bcrypt, and returns a boolean.
//
// If there is a match, or if bcrypt returns an ErrMismatchedHashAndPassword
// error, Matches returns a nil error. Otherwise, the error that occurs is
// returned.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// ValidateEmail checks whether the email provided is non-empty and valid,
// using validator.EmailRX to determine validity. If any checks fail, errors
// are added to the validator's Errors map.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(
		email,
		validator.EmailRX),
		"email",
		"must be a valid email adress",
	)
}

// ValidatePasswordPlaintext checks whether the password provided is non-empty
// and between 8 and 72 bytes long. If any checks fail, errors
// are added to the validator's Errors map.
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must be no more than 72 bytes long")
}

// ValidateUser checks various aspects of a user object. If any checks fail,
// errors are added to the validator's Errors map. Validations performed:
//
//   - Name should be non-empty and no more than 500 bytes
//   - Email should be non-empty and valid (ie, matching validator.EmailRX)
//   - Password.plaintext should be non-empty and between 8 and 72 bytes long
//
// A panic occurs if Password.hash is nil.
func ValidateUser(v *validator.Validator, u *User) {
	v.Check(u.Name != "", "name", "must be provided")
	v.Check(len(u.Name) < 500, "name", "must be no more than 500 bytes long")
	ValidateEmail(v, u.Email)

	// The plaintext password will be nil in some circumstances, so we omit the
	// validation in these cases.
	if u.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *u.Password.plaintext)
	}

	// If the plaintext password is nil, this indicates an issue with our app's
	// logic, so we panic instead of adding an error to the validation map.
	if u.Password.hash == nil {
		panic("missing password hash")
	}
}
