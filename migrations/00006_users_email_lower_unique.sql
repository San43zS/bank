-- +goose Up

UPDATE users
SET email = lower(email)
WHERE email <> lower(email);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower_unique ON users (lower(email));

-- +goose Down
DROP INDEX IF EXISTS idx_users_email_lower_unique;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);

