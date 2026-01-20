# AI Usage

This repository was implemented primarily **manually**. I used AI only for small, non-critical assistance (test ideas / sanity checks / wording), and I did not copy-paste “core business logic” from AI.

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

