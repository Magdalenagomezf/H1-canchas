# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the server
go run ./cmd/main.go

# Build binary
go build -o h1-canchas ./cmd/main.go

# Run tests
go test ./...

# Run a single package's tests
go test ./internal/service/...

# Start the database (Docker)
docker-compose up -d

# Stop the database
docker-compose down
```

There is no migration runner wired into the app — migrations in `migrations/` must be applied manually against the PostgreSQL database (e.g. with `psql`). Run them in order (001 → 004).

## Environment

Copy `.env.example` to `.env`. Required variables: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`. `PORT` defaults to `8080`.

## Architecture

The app is a REST API (Gin) for booking sports courts ("canchas") and event spaces. Entry point is `cmd/main.go`, which wires dependencies manually — no DI framework.

**Layer structure (dependency direction: controller → service → repository → DB):**

- `internal/domain/` — plain structs and constants (`User`, `Space`, `SpaceSlot`, `Booking`, role/status constants). No logic here.
- `internal/repository/` — raw SQL via `sqlx`. Repositories return `nil, nil` when a row is not found; services are responsible for turning that into `ErrNotFound`.
- `internal/service/` — business logic. Each service defines the repo interface it needs (dependency inversion). `service/errors.go` is the single source of truth for all domain errors.
- `internal/controllers/` — HTTP handlers using Gin. Parse requests, call services, map `service.Err*` to HTTP status codes.
- `internal/dto/` — request/response structs used only in the controller layer.
- `internal/middleware/` — `Auth(jwtSecret)` validates the Bearer token and sets `user_id`/`role` in the Gin context. `RequireRole(...)` gates routes by role.
- `config/` — loads `.env` into a `Config` struct passed to everything that needs it.
- `pkg/` — shared utilities: `pkg/jwt.go` (token generation/parsing), `pkg/database/postgres.go` (DB connection).

**Auth model:** JWT, HS256, 24-hour expiry. Claims carry `user_id` and `role`. Three roles: `customer`, `receptionist`, `admin`.

**Booking model:** A `Booking` can be created by a registered user (`CustomerUserID` set) or manually by staff (`CustomerUserID` nil, `CustomerName`/`CustomerPhone` set). A unique partial index on `(space_id, slot_id, booking_date)` WHERE `status IN ('pending','confirmed')` enforces availability at the DB level. Services also check availability before inserting, but the DB index is the final guard.

**Route access summary:**
- Public: `GET /spaces`, `GET /spaces/:id`, `GET /spaces/:id/slots`
- Customer: `POST /bookings`, `GET /bookings`, `GET /bookings/:id`, `PATCH /bookings/:id/cancel`
- Receptionist: `POST /spaces`, `PUT /spaces/:id`, `POST /bookings/manual`
- Admin: all of the above + `POST /admin/users`, `POST /spaces/:id/slots`, `DELETE /spaces/:id`

## Project state

For full context on implemented endpoints, deferred decisions, design decisions, bugs fixed, and what's pending for the frontend phase, read `RESUMEN-PROYECTO.md`.
