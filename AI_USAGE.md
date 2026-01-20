# AI Usage

This repository was implemented primarily **manually**. I used AI only for small, non-critical assistance (test ideas / sanity checks / wording).

## AI interactions

### AI-1
- **Purpose**: Quickly list edge-cases / scenarios to cover with unit tests for transfers/exchange (validation, insufficient funds, same-currency exchange, rounding).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Suggest a compact checklist of unit-test scenarios for a mini-banking API transfer and currency exchange flow (Go). Focus on edge cases: validation, insufficient funds, concurrency/race considerations, rounding, and idempotency assumptions.”

- **How the response was used**: Used as a checklist to verify I didn’t miss obvious cases. Tests and assertions were written manually and adapted to this codebase.

### AI-2
- **Purpose**: Sanity-check HTTP error format and status codes for common API failures (validation, auth, not found, rate limit).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Given a REST API in Go, propose a simple, consistent error response format and mapping to HTTP status codes for: validation errors, unauthorized/forbidden, not found, conflict, and rate limit.”

- **How the response was used**: Used as reference while reviewing handlers; final error shapes/statuses were chosen to match the existing project patterns and requirements.

### AI-3
- **Purpose**: Improve wording/structure of short documentation sections (README snippets / run instructions) for clarity.
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Rewrite these setup instructions to be short and clear for a reviewer (Docker compose up, migrations, running tests). Keep it concise and technical.”

- **How the response was used**: Used only as inspiration for phrasing; final documentation was edited manually to match the actual commands and repository layout.

### AI-4
- **Purpose**: Draft a minimal React routing/page structure (Vite + TS) to match backend endpoints.
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Propose a minimal React (Vite + TS) page/route structure for a banking UI (login, register, dashboard, transfer, exchange, transactions). Keep it simple and reviewer-friendly.”

- **How the response was used**: Used as a checklist. Final routes/pages were implemented manually and adapted to this repository.

### AI-5
- **Purpose**: Sanity-check “protected routes” approach (redirect unauthenticated users to login).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “In React Router, what’s a clean pattern for protected routes with auth context (redirect to /login, preserve return path)?”

- **How the response was used**: Used as reference while implementing route guards manually.

### AI-6
- **Purpose**: Sketch a minimal `AuthContext` API surface (login/register/logout, loading, error state).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Suggest a small AuthContext shape for a React + TS app (user, access token, login/register/logout methods, initial bootstrap). Keep it minimal.”

- **How the response was used**: Used as a starting point; the final context state/actions were written manually.

### AI-7
- **Purpose**: Sanity-check token storage options (in-memory vs localStorage) for a demo UI.
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “For a demo React app, what are the tradeoffs of storing access tokens in memory vs localStorage? Provide a pragmatic recommendation for a technical task.”

- **How the response was used**: Used as a decision aid; the chosen approach was implemented manually.

### AI-8
- **Purpose**: Draft a tiny API client wrapper around `fetch` (base URL, JSON parsing, typed errors).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Generate a small TypeScript `fetch` wrapper for JSON APIs: baseURL, `Authorization` header support, consistent error object from non-2xx responses.”

- **How the response was used**: Used as reference; the final client was implemented manually to match backend error responses.

### AI-9
- **Purpose**: Clarify how to map backend validation errors to a simple UI error rendering pattern.
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Given a JSON API that returns validation errors, suggest a minimal UI pattern to display field-level and form-level errors in React.”

- **How the response was used**: Used as a checklist; final error rendering was implemented manually.

### AI-10
- **Purpose**: Improve small UX details for auth forms (loading states, disable submit, error reset).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “List small UX improvements for login/register forms (loading, disabling buttons, clearing errors on change, keyboard submit). Keep it practical.”

- **How the response was used**: Used as a sanity checklist during manual implementation.

### AI-11
- **Purpose**: Sanity-check money formatting rules for UI (minor units, rounding, displaying currency).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “For a banking UI, what’s a safe/simple way to format money amounts (minor units, rounding, fixed decimals) in TypeScript? Any common pitfalls?”

- **How the response was used**: Used to validate formatting approach; final formatting utility code was written manually.

### AI-12
- **Purpose**: Draft a compact empty/loading/error state pattern for tables and dashboards.
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Suggest a minimal pattern for loading/empty/error states for a dashboard + table page in React. Keep components small.”

- **How the response was used**: Used as reference; final states/components were implemented manually.

### AI-13
- **Purpose**: Checklist for transaction history pagination UI (Prev/Next, page size, total, disabled states).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Propose a compact pagination UI checklist for a transactions table (prev/next, page size, disabled states, query params).”

- **How the response was used**: Used as a checklist; the pagination UI logic was implemented manually.

### AI-14
- **Purpose**: Sanity-check filter UX for transactions (query string syncing, reset, debouncing optional).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “What’s a clean pattern to keep filters in sync with URL query params in a React page (transactions list)? Include reset behavior.”

- **How the response was used**: Used as a reference while implementing query-param handling manually.

### AI-15
- **Purpose**: Validate basic client-side validation for transfer/exchange forms (required fields, positive amounts, same-account checks).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “List minimal client-side validation rules for transfer/exchange forms in a banking UI (required fields, >0 amount, from/to constraints).”

- **How the response was used**: Used as a checklist; validation was implemented manually.

### AI-16
- **Purpose**: Reduce repeated Tailwind `className` usage by introducing small shared UI primitives.
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “Suggest a small set of reusable UI primitives (Button, Input, Card, ErrorBox, Layout) and how to refactor pages to be more compact.”

- **How the response was used**: Used as a refactoring guide; the final `styled-components` UI primitives and page refactors were implemented manually.

### AI-17
- **Purpose**: Sanity-check small accessibility details for form components (labels, aria, error association).
- **Tool & Model**: Cursor Chat (GPT-5.2)
- **Prompt**:

  “For simple React forms, list a few high-impact accessibility checks (labels, aria-describedby for errors, focus order). Keep it minimal.”

- **How the response was used**: Used as a checklist; final markup was written manually.

