package data

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	validator "github.com/kvnloughead/greenlight/internal"
)

// Type Scope is a string type for token scopes. Valid scopes are Activation
// and Authentication, and validitiy can be checked via the Valid method.
type Scope string

const (
	Activation     Scope = "activation"
	Authentication Scope = "authentication"
)

// Returns true if the scope is valid. Valid scopes are Activation and
// Authentication.
func (s Scope) Valid() bool {
	switch s {
	case Activation, Authentication:
		return true
	default:
		return false
	}
}

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     Scope
}

// The generateToken function accepts a user ID, an expiry duration, and a
// scope, and returns a Token struct.
//
// The plaintext token is generated via cryptographically-secure pseudo-random
// generation (CSPRNG) and encoded to a base-32 string. The resulting plaintex
// string will be 26 bytes long.
//
// The hash is generated from the plaintext token using SHA-256.
func generateToken(userID int64, ttl time.Duration, scope Scope) (*Token, error) {
	token := Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Fill a slice of bytes with random bytes from CSPRNG.
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to base-32 encoded string, without padding, and store
	// in the token's Plaintext field.
	token.Plaintext = base32.StdEncoding.
		WithPadding(base32.NoPadding).
		EncodeToString(randomBytes)

	// Generate SHA-256 has of plaintext token string. The sha256.Sum256 function
	// returns an array of bytes, so we convert it to a slice for convenience.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return &token, nil
}

// ValidateTokenPlaintext uses validator.Validator to check if the plaintext
// string provided is exactly 26 bytes long. This is the number of bytes
// generated with 16 bytes of randomness are encoded into base-32.
func ValidateTokenPlaintext(v *validator.Validator, plaintext string) {
	v.Check(plaintext != "", "token", "must be provided")
	v.Check(len(plaintext) == 26, "token", "must be 26 bytes long")
}

// The TokenModel struct encapsulates database interactions with the tokens
// table.
type TokenModel struct {
	DB *sql.DB
}

// The TokenModel's New method creates a new token struct, inserts the
// corresponding record into the tokens table, and returns the token.
//
// It calls generateToken to generate the random plaintext string and its hash,
// and calls TokenModel.Insert to insert the record.
func (m TokenModel) New(userID int64, ttl time.Duration, scope Scope) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	m.Insert(token)
	return token, nil
}

// The TokenModel's Insert method adds a new record to the tokens table. It
// accepts a pointer to a Token struct and runs an INSERT query.
func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// The TokenModel's DeleteAllForUser method deletes all tokens that match
// the given scope and user ID.
func (m TokenModel) DeleteAllForUser(scope Scope, userID int64) error {

	query := `DELETE FROM tokens WHERE scope = $1 AND user_id = $2`

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

// CalculateHash takes a string a returns its SHA-256 hash.
func CalculateHash(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}
