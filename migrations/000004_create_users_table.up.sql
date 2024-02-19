CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  name text NOT NULL,
  email citext UNIQUE NOT NULL, --- citext: case-insensitive text
  password_hash bytea NOT NULL, --- bytea: binary string (for bcrypt hash)
  activated bool NOT NULL, --- will be false until user confirms email 
  version integer NOT NULL DEFAULT 1
);