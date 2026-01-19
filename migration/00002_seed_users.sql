-- +goose Up

-- User 1
INSERT INTO users (id, email, password, first_name, last_name, created_at, updated_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'user1@test.com',
    '$2a$10$MDqECt0NP5tsJXlvQo5.wubSUQV5I7GdLv7CBj/7szgLI4yVopkyi', -- password123
    'John',
    'Doe',
    NOW(),
    NOW()
);

INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'USD', 1000.00, NOW(), NOW()),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'EUR', 500.00, NOW(), NOW());

-- User 2
INSERT INTO users (id, email, password, first_name, last_name, created_at, updated_at)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'user2@test.com',
    '$2a$10$MDqECt0NP5tsJXlvQo5.wubSUQV5I7GdLv7CBj/7szgLI4yVopkyi', -- password123
    'Jane',
    'Smith',
    NOW(),
    NOW()
);

INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
VALUES
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222222', 'USD', 1000.00, NOW(), NOW()),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222222', 'EUR', 500.00, NOW(), NOW());

-- User 3
INSERT INTO users (id, email, password, first_name, last_name, created_at, updated_at)
VALUES (
    '33333333-3333-3333-3333-333333333333',
    'user3@test.com',
    '$2a$10$MDqECt0NP5tsJXlvQo5.wubSUQV5I7GdLv7CBj/7szgLI4yVopkyi', -- password123
    'Bob',
    'Johnson',
    NOW(),
    NOW()
);

INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
VALUES
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '33333333-3333-3333-3333-333333333333', 'USD', 1000.00, NOW(), NOW()),
    ('ffffffff-ffff-ffff-ffff-ffffffffffff', '33333333-3333-3333-3333-333333333333', 'EUR', 500.00, NOW(), NOW());

-- +goose Down
DELETE FROM accounts WHERE user_id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);

DELETE FROM users WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);
