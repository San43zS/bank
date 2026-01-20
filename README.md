# Mini Banking Platform (Go)

Backend service for a mini banking platform (users, accounts, transfers, currency exchange) with PostgreSQL storage and SQL migrations.

## Requirements

- Go \(for local development\): Go 1.21+
- Docker + Docker Compose \(recommended for running the full stack\)

## Quick start (Docker, recommended)

From the `technical task/` directory:

```bash
docker compose up --build
```

This starts:
- `postgres` on `localhost:5433`
- `migrate` \(runs migrations automatically on startup\)
- `api` on `localhost:8080`

Stop everything:

```bash
docker compose down
```

Reset the database (removes the Docker volume):

```bash
docker compose down -v
```

## Configuration (environment variables)

The service reads configuration from environment variables (optionally from a local `.env` file).

- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `5433`)
- `DB_USER` (default: `postgres`)
- `DB_PASSWORD` (default: `postgres`)
- `DB_NAME` (default: `banking`)
- `PORT` (default: `8080`)
- `JWT_SECRET` (default: `bank`)
- `CONSISTENCY_CRON_ENABLED` (default: `false`)
- `CONSISTENCY_CRON_INTERVAL_SECONDS` (default: `300`)
- `CONSISTENCY_CRON_TIMEOUT_SECONDS` (default: `30`)
- `RATE_LIMIT_ENABLED` (default: `false`)
- `RATE_LIMIT_RPS` (default: `10`)
- `RATE_LIMIT_BURST` (default: `20`)
- `EXCHANGE_RATE_USD_TO_EUR` (default: `0.92`)

## Migrations

### Run migrations via Docker (no local tools needed)

```bash
docker compose run --rm migrate
```

### Run migrations locally (optional)

The `Makefile` uses `goose`. If you do not have it installed, install it using your OS package manager, or use the Docker command above.

To apply migrations against the local Postgres started by compose (`localhost:5433`):

```bash
make migrate-up
```

Other useful targets:

```bash
make migrate-status
make migrate-down
```

## Local development (without Docker for the API)

1. Start Postgres:

```bash
docker compose up -d postgres
```

2. Run migrations:

```bash
docker compose run --rm migrate
```

3. Start the API:

```bash
go run ./cmd/app
```

The API will be available on `http://localhost:8080`.

## API documentation

### Authentication

Protected endpoints require:

```
Authorization: Bearer <access_token>
```

### Error format

Generic errors:

```json
{ "error": "message" }
```

Validation errors (bad request body / binding errors):

```json
{
  "error": "validation_error",
  "fields": [
    { "field": "email", "message": "must be a valid email" }
  ]
}
```

Rate limit errors:

```json
{ "error": "rate_limited" }
```

### Endpoints

- `POST /auth/register` — Register a new user.

Request body:

```json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

- `POST /auth/login` — Login and receive access/refresh tokens.

Request body:

```json
{
  "email": "user1@test.com",
  "password": "password123"
}
```

- `POST /auth/refresh` — Refresh access token using refresh token.
- `POST /auth/logout` — Revoke refresh token.
- `GET /auth/me` — Get current authenticated user info.
- `GET /accounts` — List current user accounts.
- `GET /accounts/:id/balance` — Get balance for a specific account.
- `POST /transactions/transfer` — Transfer money to another user (by user ID or email).

Request body (by user id):

```json
{
  "to_user_id": "22222222-2222-2222-2222-222222222222",
  "currency": "USD",
  "amount": 10.25
}
```

Request body (by email):

```json
{
  "to_user_email": "user2@test.com",
  "currency": "USD",
  "amount": 10.25
}
```

- `POST /transactions/exchange` — Exchange between USD/EUR using fixed rate.
- `GET /transactions` — List current user transactions.
- `GET /health` — Health check.

## Seeded test users (from migrations)

After running migrations, the database is seeded with the following users:
- `user1@test.com`
- `user2@test.com`
- `user3@test.com`

Password: `password123`

## Tests

```bash
go test ./...
```

