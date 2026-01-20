# Mini Banking Platform

Repository structure:

- `backend/` — Go backend + PostgreSQL migrations + OpenAPI
- `frontend/` — React (Vite) UI

## Requirements

- Go 1.21+
- Node.js 19+ (tested with Node 24)
- Docker + Docker Compose

## Quick start (Docker backend + local frontend)

### Backend

From repo root:

```bash
cd backend
docker compose up --build
```

This starts:
- `postgres` on `localhost:5433`
- `migrate` (runs migrations on startup)
- `api` on `localhost:8080`

Stop everything:

```bash
cd backend
docker compose down
```

### Frontend

From repo root:

```bash
cd frontend
npm install
npm run dev
```

The UI will be available at the Vite URL (usually `http://localhost:5173`).

## Configuration

### Backend environment variables

Backend reads configuration from env vars (optionally from a local `.env` file in `backend/`):

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
- `CRON_STOP_TIMEOUT_SECONDS` (default: `5`)
- `SHUTDOWN_TIMEOUT_SECONDS` (default: `10`)
- `RATE_LIMIT_ENABLED` (default: `false`)
- `RATE_LIMIT_RPS` (default: `10`)
- `RATE_LIMIT_BURST` (default: `20`)
- `EXCHANGE_RATE_USD_TO_EUR` (default: `0.92`)

### Frontend environment variables

Frontend uses Vite env var:

- `VITE_API_BASE_URL` (default: `http://localhost:8080`)

Example file is in `frontend/env.example`.

## API docs

- OpenAPI spec: `backend/docs/openapi.yaml`
- How to view: `backend/docs/README.md`

## Tests

Backend:

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend
npm run build
```

## Seeded test users (from migrations)

After running migrations, the database is seeded with:
- `user1@test.com`
- `user2@test.com`
- `user3@test.com`

Password: `password123`

