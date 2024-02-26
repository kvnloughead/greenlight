--- The tokens table contains hashes of activation and authentication tokens.

CREATE TABLE IF NOT EXISTS tokens (
  --- A SHA-256 hash of the activation token.
  hash bytea PRIMARY KEY,

  --- A reference to a record from users table. DELETE CASCADE ensures that all
  --- tokens associated with a user are deleted if the user record is deleted.
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,

  expiry timestamp(0) with time zone NOT NULL,

  --- Scopes include activation and authentication.
  scope text NOT NULL
);