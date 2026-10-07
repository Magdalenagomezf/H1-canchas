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

# Build binary
cd backend && go build -o h1-canchas ./cmd/main.go

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
```

There is no migration runner wired into the app — migrations in `backend/migrations/` must be applied manually against the PostgreSQL database (e.g. with `psql`). Run them in order (001 → 009).

## Environment

**Backend:** copy `backend/.env.example` to `backend/.env`.

- Required: `JWT_SECRET`, `MP_ACCESS_TOKEN`, `MP_WEBHOOK_SECRET`, `MP_WEBHOOK_URL`.
- Database: either `DATABASE_URL` (full connection string, takes precedence) or `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`.
- Optional (with defaults): `PORT` (`8080`), `ALLOWED_ORIGIN` (`http://localhost:5173`), `BOOKING_HOLD_TTL_MINUTES` (`20`), `DEPOSIT_PERCENTAGE` (`0.15`), `BOOKING_EXPIRY_SWEEP_INTERVAL_MINUTES` (`2`), `FRONTEND_URL` (`http://localhost:5173`, used for links in emails), `EMAIL_FROM` (`H1 Canchas <onboarding@resend.dev>`), `RESEND_API_KEY` (default empty: emails are only logged — recipient and subject — and never sent).

**Frontend:** `frontend/.env` already contains `VITE_API_URL=http://localhost:8080`. Adjust if needed.

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
- `pkg/` — shared utilities: `pkg/jwt.go` (token generation/parsing), `pkg/database/postgres.go` (DB connection), `pkg/mercadopago/` (Checkout Pro client, preferences, payment lookup, webhook signature verification).

**Auth model:** JWT, HS256, 24-hour expiry. Claims carry `user_id` and `role`. Three roles: `customer`, `receptionist`, `admin`.

**Booking model:** A `Booking` can be created by a registered user (`CustomerUserID` set) or manually by staff (`CustomerUserID` nil, `CustomerName`/`CustomerPhone` set). A unique partial index on `(space_id, slot_id, booking_date)` WHERE `status IN ('pending','confirmed')` enforces availability at the DB level. Services also check availability before inserting, but the DB index is the final guard.

Online bookings start as `pending` with a hold (`expires_at`, default 20 min) to pay the deposit. A background ticker in `cmd/main.go` moves stale holds to `expired`. Payment is split into two independent parts: `deposit_status` and `balance_status`.

**Payments (Mercado Pago Checkout Pro):** the webhook never trusts the notification payload — it always re-fetches the payment from the MP API, which is the real security barrier. `x-signature` is checked but a mismatch only logs a WARN (Checkout Pro sends legacy IPN notifications whose signature never validates). Handling is idempotent.

**Booking batches:** staff can create recurring weekly bookings (`recurring_teacher`) or maintenance blocks (`maintenance`). Both materialize real `confirmed` bookings tagged with `batch_id` (max range 366 days); already-taken dates are skipped. Cancelling a batch cancels only today's and future bookings.

**Route access summary:**
- Public: `GET /health`, `GET /spaces`, `GET /spaces/:id`, `GET /spaces/:id/slots`, `POST /payments/webhook`
- Any authenticated user: `POST /bookings`, `GET /bookings` (own bookings), `GET /bookings/:id`, `PATCH /bookings/:id/cancel`, `POST /bookings/:id/payments`
- Receptionist: `PUT /spaces/:id`, `GET /admin/bookings`, `POST /bookings/manual`, `POST /bookings/recurring`, `POST /bookings/block`, `GET /bookings/batches`, `PATCH /bookings/batches/:id/cancel`, `PATCH /bookings/:id/payment`
- Admin: all of the above + `POST /spaces`, `POST /spaces/:id/slots`, `DELETE /spaces/:id`, `DELETE /admin/spaces/:id`, `GET/POST /admin/users`, `PATCH /admin/users/:id/role`, `DELETE /admin/users/:id`

Only admins can create spaces — receptionists can edit them but not create them (business decision).

## Project state

For full context on implemented endpoints, deferred decisions, design decisions, bugs fixed, and what's pending for the frontend phase, read `RESUMEN-PROYECTO.md`.
