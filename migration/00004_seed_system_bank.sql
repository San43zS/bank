-- +goose Up

INSERT INTO users (id, email, password, first_name, last_name, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'bank@system.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- password123 (unused)
    'System',
    'Bank',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', 'USD', 1000000000.00, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000001', 'EUR', 1000000000.00, NOW(), NOW())
ON CONFLICT (user_id, currency) DO NOTHING;

-- +goose Down
