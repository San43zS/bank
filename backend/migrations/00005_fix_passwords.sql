-- +goose Up

UPDATE users 
SET password = '$2a$10$MDqECt0NP5tsJXlvQo5.wubSUQV5I7GdLv7CBj/7szgLI4yVopkyi'
WHERE email IN ('user1@test.com', 'user2@test.com', 'user3@test.com');

-- +goose Down
UPDATE users 
SET password = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
WHERE email IN ('user1@test.com', 'user2@test.com', 'user3@test.com');
