# Mini Banking Platform

A full-stack mini banking platform with a **Go backend** and **React 19 (Vite) frontend**. The backend implements **double-entry bookkeeping** (ledger as an audit trail) and maintains **cached account balances** for fast reads while enforcing consistency.

## Table of Contents

- [Setup Instructions](#setup-instructions)
- [User Management Approach](#user-management-approach)
- [Architecture Overview](#architecture-overview)
- [Double-Entry Ledger Design](#double-entry-ledger-design)
- [Maintaining Balance Consistency](#maintaining-balance-consistency)
- [Design Decisions and Trade-offs](#design-decisions-and-trade-offs)
- [Known Limitations](#known-limitations)
- [Incomplete Features Due to Time Constraints](#incomplete-features-due-to-time-constraints)
- [Technical Questions Answered](#technical-questions-answered)
- [Technical Evaluation Checklist](#technical-evaluation-checklist)
- [API Documentation](#api-documentation)
- [AI / LLM Usage](#ai--llm-usage)

---

## Setup Instructions

### Prerequisites

- **Docker** and **Docker Compose**
- **Go 1.21+** (for local backend development)
- **Node.js 20+** and **npm** (for local frontend development)
- **goose** CLI (only if running migrations locally)
- `make` (optional; convenience wrapper on Unix/macOS)

### Quick Start with Docker (Recommended)

1. **Clone the repository**

```bash
git clone <repository-url>
cd <repository-folder>
```

2. **Start backend services (PostgreSQL + migrations + API)**

```bash
cd backend
docker compose up --build
```

This starts:
- PostgreSQL on `localhost:5433`
- `migrate` service (runs goose migrations on startup)
- API on `http://localhost:8080`

Stop backend services:

```bash
cd backend
docker compose down
```

3. **Start frontend (local)**

```bash
cd ../frontend
npm install
npm run dev
```

Frontend will be available at `http://localhost:5173` (Vite default).

### Configuration

#### Backend environment variables

Backend reads env vars (optionally from a local `.env` file in `backend/`).

- `DB_HOST` (default: `localhost`) — local run; in Docker uses `postgres`
- `DB_PORT` (default: `5433`) — local run; in Docker uses `5432`
- `DB_USER` (default: `postgres`)
- `DB_PASSWORD` (default: `postgres`)
- `DB_NAME` (default: `banking`)
- `PORT` (default: `8080`)
- `JWT_SECRET` (default: `bank`)
- `EXCHANGE_RATE_USD_TO_EUR` (default: `0.92`)
- `CONSISTENCY_CRON_ENABLED` (default: `false`) — periodic consistency checks (useful for review)
- `CONSISTENCY_CRON_INTERVAL_SECONDS` (default: `300`)
- `CONSISTENCY_CRON_TIMEOUT_SECONDS` (default: `30`)
- `RATE_LIMIT_ENABLED` (default: `false`) — in-memory IP rate limiting
- `RATE_LIMIT_RPS` (default: `10`)
- `RATE_LIMIT_BURST` (default: `20`)

#### Frontend environment variables

Frontend uses Vite env var:
- `VITE_API_BASE_URL` (default: `http://localhost:8080`)

Example file: `frontend/env.example`.

### Local Development Setup (without Docker API container)

```bash
cd backend
docker compose up -d postgres

# Option A: using Makefile (if available)
make migrate-up
make run

# Option B: without make
goose -dir ./migrations postgres "postgres://postgres:postgres@localhost:5433/banking?sslmode=disable" up
go run ./cmd/app
```

### Running Tests

```bash
cd backend
go test ./...
```

Frontend build (sanity check):

```bash
cd frontend
npm run build
```

---

## User Management Approach

### Chosen approach: Registration (Option A)

The backend provides `POST /auth/register`.

How it works:
1. User submits email/password + first/last name
2. Password is hashed (bcrypt)
3. USD/EUR accounts are created
4. Initial balances are funded via **ledger-backed transfers** from a seeded system bank user:
   - USD: **$1000.00**
   - EUR: **€500.00**
5. Access + refresh tokens are returned

### Seeded demo users (for reviewer convenience)

Migrations also seed 3 demo users (so you can test immediately without registering):
- `user1@test.com`
- `user2@test.com`
- `user3@test.com`

Password: `password123`

---

## Architecture Overview

### Backend (Go)

```
backend/
├── cmd/app/                 # application entrypoint
├── config/                  # env config loading
├── internal/
│   ├── cron/                # periodic consistency checks (optional)
│   ├── domain/              # domain types + money helpers (int64 cents)
│   ├── http/                # Gin handlers, DTOs, middleware
│   ├── repo/                # PostgreSQL repositories + tx runner
│   └── service/             # orchestration services (auth/accounts/transactions)
├── migrations/              # goose migrations (schema + seed data)
└── docs/openapi.yaml        # OpenAPI specification
```

### Frontend (React 19 + Vite)

```
frontend/src/
├── pages/                   # Login/Register/Dashboard/Transfer/Exchange/Transactions
├── auth/                    # AuthContext
├── api/                     # fetch wrapper + API types
├── lib/                     # money helpers (decimal <-> cents)
└── ui/                      # small UI primitives (Tailwind + minimal styling)
```

---

## Double-Entry Ledger Design

### Core concept

Every financial operation writes immutable rows to `ledger`. Each transaction produces entries that are balanced:

- Transfer: 2 entries (debit/credit) in the same currency.
- Exchange: 4 entries (two per currency leg) using the system bank as counterparty.

Because currency is defined by the account (`accounts.currency`), the ledger is effectively balanced per currency-leg.

### Schema

See `backend/migrations/00001_init_schema.sql`. Minimal tables:
- `users`
- `accounts` (USD/EUR per user, cached balance)
- `transactions` (user-facing history)
- `ledger` (audit trail; positive/negative amounts)

### Examples

**Transfer $50 from User A to User B (USD)**

| account | amount |
|---|---:|
| User A USD | -50.00 |
| User B USD | +50.00 |
| **Sum** | **0.00** |

**Exchange $100 USD → €92 EUR (rate 0.92)**

| account | amount |
|---|---:|
| User USD | -100.00 |
| System bank USD | +100.00 |
| System bank EUR | -92.00 |
| User EUR | +92.00 |
| **Sum** | **0.00** |

---

## Maintaining Balance Consistency

### Atomicity

Transfers and exchanges execute inside a single DB transaction (`WithTx`):
- lock required accounts (`SELECT ... FOR UPDATE`)
- insert into `transactions`
- insert corresponding `ledger` entries
- verify ledger balance for the transaction
- update cached balances in `accounts`

Any error ⇒ rollback (no partial updates).

### Concurrency / double-spending

- Uses row-level locks (`FOR UPDATE`)
- Locks are acquired in deterministic order (sorted UUIDs) to reduce deadlocks
- Insufficient funds are checked after locking the balance row

### Consistency checks

There are two layers:
1. **Per-transaction ledger check**: verify that a transaction’s ledger entries sum to 0.
2. **Account vs ledger reconciliation check** (optional cron): detect accounts where `accounts.balance != SUM(ledger.amount)`.

Enable cron checks with:

```bash
CONSISTENCY_CRON_ENABLED=true
```

### Currency precision

- Money is represented in application as **int64 cents**.
- DB stores `DECIMAL(15,2)` and values are written as decimal strings derived from cents.
- Exchange conversion uses integer arithmetic with rounding to cents.

---

## Design Decisions and Trade-offs

1) **Cached balances in `accounts`**
- **Why**: fast reads for dashboard/history without aggregating ledger on every request
- **Trade-off**: requires strong invariants + periodic checks

2) **Ledger is the audit trail; balances are a cache**
- **Why**: ledger is append-only; balances are updated transactionally for performance
- **Trade-off**: application must enforce invariants (no DB triggers in this implementation)

3) **System bank as exchange counterparty**
- **Why**: keeps cross-currency exchange balanced using a 4-entry pattern
- **Trade-off**: introduces a system user with “liquidity” constraints

4) **Equity / Opening balances postings**
- **Why**: for seeded/demo data, opening balances must also be explainable in double-entry terms
- **Trade-off**: requires a one-time reconciliation posting for older DBs (see migration `00007`)

---

## Known Limitations

1) **No idempotency keys**
- Re-sending the same request creates a new transaction.

2) **CORS is wide open**
- `Access-Control-Allow-Origin: *` for demo convenience.

3) **Rate limiter is in-memory (optional)**
- Not shared across instances.

4) **Transactions list is list-only pagination**
- No total count / cursor pagination.

---

## Incomplete Features Due to Time Constraints

- **Reconciliation endpoint**: there is a consistency-check cron, but no `/system/reconcile` API.
- **Real-time updates**: no WebSockets.
- **Receipts/details modal**: not implemented.
- **Admin/audit UI**: ledger exists in DB, no admin UI.

---

## Technical Questions Answered

### How do you ensure transaction atomicity?
All write operations (transaction row, ledger entries, balance updates) are executed in a single SQL transaction (`WithTx`). Any error triggers rollback.

### How do you prevent double-spending?
Accounts involved in transfer/exchange are locked with `SELECT ... FOR UPDATE`, then balance checks happen on the locked rows.

### How do you maintain consistency between ledger entries and account balances?
Balances are updated only after ledger entries are written and the transaction is verified as balanced. Optional cron checks detect mismatches (`accounts.balance != SUM(ledger)`).

### How do you handle decimal precision?
Application uses int64 cents; DB stores `DECIMAL(15,2)`; exchange uses integer rounding to cents.

### What indexing strategy is used?
Indexes from `00001_init_schema.sql` include:
- `idx_ledger_transaction_id`, `idx_ledger_account_id`, `idx_ledger_created_at`
- `idx_transactions_created_at`, `idx_transactions_from_account`, `idx_transactions_to_account`
- `idx_accounts_user_id`, `idx_accounts_currency`

### How do you verify balances are correctly synchronized?
Enable `CONSISTENCY_CRON_ENABLED=true` and observe logs:
- “Ledger consistency check OK/FAILED”
- “Account balance consistency check OK/FAILED”

### How would you scale this system?
Typical path:
- connection pooling (PgBouncer), better pagination, proper read models for history
- partition ledger/transactions by time, archive old entries
- separate read replicas for history, keep writes on primary

---

## Technical Evaluation Checklist

### Must Have (70%)

- [x] Functional double-entry ledger implementation
- [x] Ledger entries maintained as audit trail
- [x] Account balances kept in sync with ledger (transactionally + cron verification)
- [x] Transfer + exchange implemented
- [x] Concurrency handling (`SELECT ... FOR UPDATE`, deterministic lock order)
- [x] Prevention of invalid states (insufficient funds, currency checks)
- [x] Authentication working (JWT)
- [x] Functional UI (login/register, dashboard, transfer, exchange, transactions)

### Should Have (20%)

- [x] Clear error responses + frontend error display
- [x] Loading states in UI
- [x] Database migrations/seed data (goose)
- [x] Environment configuration
- [x] Input validation (backend; basic client-side checks)
- [ ] Transaction confirmation step (not implemented)

### Nice to Have (10%)

- [x] OpenAPI documentation (`backend/docs/openapi.yaml`)
- [x] Docker setup (Compose)
- [x] Optional consistency checks cron
- [ ] Reconciliation API endpoint (`/system/reconcile`)
- [ ] WebSocket real-time updates
- [ ] Receipts/details modal

---

## API Documentation

- OpenAPI spec: `backend/docs/openapi.yaml`

Key endpoints:

| Method | Endpoint | Description |
|---|---|---|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login |
| GET | `/auth/me` | Current user |
| GET | `/accounts` | List accounts |
| GET | `/accounts/:id/balance` | Account balance |
| POST | `/transactions/transfer` | Transfer (same currency) |
| POST | `/transactions/exchange` | Exchange (USD/EUR) |
| GET | `/transactions` | History (filter + pagination) |

---

## AI / LLM Usage

See `AI_USAGE.md`.

