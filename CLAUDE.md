# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project structure

```
H1-canchas/
├── backend/      # Go REST API (Gin)
│   ├── cmd/      # Entry point
│   ├── config/
│   ├── internal/ # controllers, service, repository, domain, middleware, dto
│   ├── migrations/
│   └── pkg/      # jwt, database
├── frontend/     # React + TypeScript + Vite
│   └── src/
├── docker-compose.yml
└── CLAUDE.md
```

## Commands

All backend commands must be run from the `backend/` directory.

```bash
# Run the server
cd backend && go run ./cmd/main.go

# Build binary (build the package ./cmd, not ./cmd/main.go: only then Go embeds
# the git commit that GET /health reports)
cd backend && go build -o h1-canchas ./cmd

# Smoke test a running backend (optionally checking the commit it reports)
bash scripts/smoke.sh http://localhost:8080 $(git rev-parse --short HEAD)

# Run tests
cd backend && go test ./...

# Run a single package's tests
cd backend && go test ./internal/service/...

# Start the database (Docker) — run from project root
docker-compose up -d

# Stop the database
docker-compose down

# Normalize existing phone numbers to E.164 (dry run; add --apply to write).
# Required right after deploying the phase-1 backend, see RESUMEN-PROYECTO.md.
cd backend && go run ./cmd/normalize-phones

# Frontend dev server
cd frontend && npm run dev

# Frontend build
cd frontend && npm run build

# Frontend unit tests (Vitest; pure logic in src/lib, tests next to the code as *.test.ts)
cd frontend && npm test

# API integration tests (Playwright without a browser, frontend/e2e/api.spec.ts): status codes,
# role checks and the JSON contract. Same variables as the E2E, without E2E_BASE_URL.
cd frontend && E2E_API_URL=http://localhost:8080 E2E_ADMIN_PHONE=<admin phone> E2E_ADMIN_PASSWORD=<admin password> npm run test:api

# E2E tests (Playwright, frontend/e2e). The app must be running; each test creates and
# cleans its own data through an admin account. First time: npx playwright install chromium
cd frontend && E2E_BASE_URL=http://localhost:5173 E2E_API_URL=http://localhost:8080 \n  E2E_ADMIN_PHONE=<admin phone> E2E_ADMIN_PASSWORD=<admin password> npm run e2e
```

CI (`.github/workflows/ci.yml`) runs on every PR and push to `main`: backend `go vet`, `go build` and `go test -race` against a Postgres 16 service (with `REQUIRE_DB=1`, so the repository integration tests fail instead of skipping when the database is missing), a smoke test (`scripts/smoke.sh`) against a fresh build that must report the PR's commit, frontend `npm run lint`, `npm test` and `npm run build`, and, against a throwaway stack, the API integration tests followed by the Playwright E2E tests (Postgres + backend + `vite preview`, seeded with `frontend/e2e/ci-seed.sql`). The E2E answer the Mercado Pago preference call themselves, because CI has no MP credentials. Run the same checks locally before pushing.

Deploys go only through `.github/workflows/deploy.yml`, triggered when CI passes on `main`: deploy to QA → smoke → API integration → E2E against QA → manual approval (GitHub environment `production`) → deploy to PROD → smoke. Each run deploys the commit CI verified (Render deploy hook with `ref=<sha>`, Vercel CLI from a checkout of that commit); provider auto-deploys are off. Setup, migrations, promotion and rollback: `docs/DEPLOY.md`.

There is no migration runner wired into the app — migrations in `backend/migrations/` must be applied manually against the PostgreSQL database (e.g. with `psql`). Run them in order (001 → 011).

## Environment

**Backend:** copy `backend/.env.example` to `backend/.env`.

- Required: `JWT_SECRET`, `MP_ACCESS_TOKEN`, `MP_WEBHOOK_SECRET`, `MP_WEBHOOK_URL`.
- Database: either `DATABASE_URL` (full connection string, takes precedence) or `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`.
- Optional (with defaults): `PORT` (`8080`), `ALLOWED_ORIGIN` (`http://localhost:5173`), `BOOKING_HOLD_TTL_MINUTES` (`20`), `DEPOSIT_PERCENTAGE` (`0.15`), `BOOKING_EXPIRY_SWEEP_INTERVAL_MINUTES` (`2`), `FRONTEND_URL` (`http://localhost:5173`, used for links in emails), `EMAIL_FROM` (`H1 Canchas <onboarding@resend.dev>`), `RESEND_API_KEY` (default empty: emails are only logged — recipient and subject — and never sent), `GOOGLE_CLIENT_ID` (OAuth Web client ID for Google login; default empty: `POST /auth/google` answers 503 and the rest of the app works).

**Frontend:** copy `frontend/.env.example` to `frontend/.env` (`VITE_API_URL`, default `http://localhost:8080`). Optional: `VITE_GOOGLE_CLIENT_ID` (same value as the backend `GOOGLE_CLIENT_ID`; empty hides the Google button). Vite reads it only at startup.

## Architecture

The app is a REST API (Gin) for booking sports courts ("canchas") and event spaces. Entry point is `backend/cmd/main.go`, which wires dependencies manually — no DI framework.

**Layer structure (dependency direction: controller → service → repository → DB):**

- `internal/domain/` — plain structs and constants (`User`, `Space`, `SpaceSlot`, `Booking`, `Payment`, `BookingBatch`, role/status constants). No logic here.
- `internal/repository/` — raw SQL via `sqlx`. Repositories return `nil, nil` when a row is not found; services are responsible for turning that into `ErrNotFound`.
- `internal/service/` — business logic. Each service defines the repo interface it needs (dependency inversion). `service/errors.go` is the single source of truth for all domain errors.
- `internal/controllers/` — HTTP handlers using Gin. Parse requests, call services, map `service.Err*` to HTTP status codes.
- `internal/dto/` — request/response structs used only in the controller layer.
- `internal/middleware/` — `Auth(jwtSecret)` validates the Bearer token and sets `user_id`/`role` in the Gin context. `RequireRole(...)` gates routes by role.
- `config/` — loads `.env` into a `Config` struct passed to everything that needs it.
- `pkg/` — shared utilities: `pkg/google/` (Google ID token verifier, via `idtoken`), `pkg/jwt.go` (token generation/parsing), `pkg/database/postgres.go` (DB connection), `pkg/mercadopago/` (Checkout Pro client, preferences, payment lookup, webhook signature verification).

**Auth model:** JWT, HS256, 24-hour expiry. Claims carry `user_id` and `role`. Three roles: `customer`, `receptionist`, `admin`.

**Google login:** `POST /auth/google` takes a Google ID token, finds the user by `google_sub`, links an existing `customer` with the same email, or creates a `customer` with `phone = NULL` and `password_hash = NULL`. Staff accounts are never linked (403). `users.phone`, `users.password_hash` and `users.google_sub` are nullable (`*string` in `domain.User`); a CHECK guarantees at least a password or a `google_sub`. Booking online requires a phone (`ErrProfilePhoneMissing`, HTTP 422 with `code: "phone_required"`); users set it with `PATCH /me`.

**Password reset:** `POST /auth/password-reset` `{email}` always answers 202 with the same body and emails the link in a background goroutine, so neither the answer nor its timing reveals whether the account exists. The link is `FRONTEND_URL/restablecer-contrasena#token=<token>` (fragment: the token never reaches server logs). Tokens are 32 bytes from `crypto/rand`; `password_reset_tokens` stores only their SHA-256, they last 1 hour, and a user gets at most one every 2 minutes. `POST /auth/password-reset/confirm` `{token, password}` answers 204, or 400 with `code: "invalid_reset_token"` for any unknown, used or expired token or inactive user; it sets the password and invalidates all of the user's pending tokens in one transaction. `POST /auth/password-reset/check` `{token}` answers 204 or the same 400 without using the token up; the reset page calls it on load so a used or expired link shows its message before the form. Both share the redeemable-token conditions (`redeemableTokenWhere`). It works for Google-only accounts too. Existing JWTs stay valid until they expire.

**Booking model:** A `Booking` can be created by a registered user (`CustomerUserID` set) or manually by staff (`CustomerUserID` nil, `CustomerName`/`CustomerPhone` set). A unique partial index on `(space_id, slot_id, booking_date)` WHERE `status IN ('pending','confirmed')` enforces availability at the DB level. Services also check availability before inserting, but the DB index is the final guard.

Online bookings start as `pending` with a hold (`expires_at`, default 20 min) to pay the deposit. A background ticker in `cmd/main.go` moves stale holds to `expired`. Payment is split into two independent parts: `deposit_status` and `balance_status`.

**Payments (Mercado Pago Checkout Pro):** the webhook never trusts the notification payload — it always re-fetches the payment from the MP API, which is the real security barrier. `x-signature` is checked but a mismatch only logs a WARN (Checkout Pro sends legacy IPN notifications whose signature never validates). Handling is idempotent.

**Booking batches:** staff can create recurring weekly bookings (`recurring_teacher`) or maintenance blocks (`maintenance`). Both materialize real `confirmed` bookings tagged with `batch_id` (max range 366 days); already-taken dates are skipped. Cancelling a batch cancels only today's and future bookings.

**Route access summary:**
- Public: `GET /health` (pings the database: 200 `{status, db, commit}`, or 503 when the database is unreachable; `commit` comes from `APP_COMMIT`, then `RENDER_GIT_COMMIT` (set by Render), then the git revision embedded at build time, `unknown` under `go run`), `GET /spaces`, `GET /spaces/:id`, `GET /spaces/:id/slots`, `POST /payments/webhook`, `POST /auth/register`, `POST /auth/login`, `POST /auth/google`, `POST /auth/password-reset`, `POST /auth/password-reset/confirm`, `POST /auth/password-reset/check`
- Any authenticated user: `GET /me`, `PATCH /me`, `POST /bookings`, `GET /bookings` (own bookings), `GET /bookings/:id`, `PATCH /bookings/:id/cancel`, `POST /bookings/:id/payments`
- Receptionist: `PUT /spaces/:id`, `GET /admin/bookings`, `POST /bookings/manual`, `POST /bookings/recurring`, `POST /bookings/block`, `GET /bookings/batches`, `PATCH /bookings/batches/:id/cancel`, `PATCH /bookings/:id/payment`
- Admin: all of the above + `POST /spaces`, `POST /spaces/:id/slots`, `DELETE /spaces/:id`, `DELETE /admin/spaces/:id`, `GET/POST /admin/users`, `PATCH /admin/users/:id/role`, `DELETE /admin/users/:id`

Only admins can create spaces — receptionists can edit them but not create them (business decision).

## Project state

For full context on implemented endpoints, deferred decisions, design decisions, bugs fixed, and what's pending for the frontend phase, read `RESUMEN-PROYECTO.md`.
