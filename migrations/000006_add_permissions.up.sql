--- The permissions table stores permission indicators.
CREATE TABLE IF NOT EXISTS permissions (
  id bigserial PRIMARY KEY,
  code text NOT NULL
);

--- The users_permissions table is a join table storing all user/permission 
--- pairs in the database. The primary key is a composite of the primary keys
--- of the two joined tables.
---
--- The REFERENCES syntax establishes foreign key constraints against the 
--- primary keys of the users and permissions tables.
CREATE TABLE IF NOT EXISTS users_permissions (
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
  PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES 
  ('movies:read'),
  ('movies:write');