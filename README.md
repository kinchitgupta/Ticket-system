# Ticket System (Backend)

A small backend service where a user can register, log in, create tickets, view only their own tickets,
and update the status of their own tickets.

## Tech

- Language: Go 1.22 (standard library `net/http` with method+path routing, no framework)
- Auth: JWT (HS256), via `github.com/golang-jwt/jwt/v5`
- Passwords: hashed with PBKDF2-HMAC-SHA256 + random per-user salt (100,000 iterations), implemented
  against the Go standard library only — never stored in plain text
- Storage: in-memory store (thread-safe, guarded by a mutex). Swappable for SQLite/Postgres later —
  handlers only depend on the `Store` interface in `internal/store`.

## Project Structure

```
cmd/main.go                        - entrypoint, routes
internal/models/models.go          - User, Ticket, status transition rules
internal/store/store.go            - in-memory data store
internal/auth/auth.go              - password hashing + JWT generation/parsing
internal/middleware/auth.go        - JWT auth middleware
internal/handlers/auth_handlers.go - /auth/register, /auth/login
internal/handlers/ticket_handlers.go - /tickets endpoints
internal/handlers/helpers.go       - shared JSON response helpers
```

## Endpoints

| Method | Endpoint                | Auth required | Purpose                    |
|--------|--------------------------|---------------|-----------------------------|
| GET    | /health                  | No            | Health check                |
| POST   | /auth/register           | No            | Register user                |
| POST   | /auth/login              | No            | Login, returns JWT           |
| POST   | /tickets                 | Yes           | Create ticket                |
| GET    | /tickets                 | Yes           | List logged-in user's tickets|
| GET    | /tickets/{id}            | Yes           | Get own ticket by ID          |
| PATCH  | /tickets/{id}/status     | Yes           | Update own ticket status      |

Protected routes require header: `Authorization: Bearer <token>`

### Status flow

```
open -> in_progress -> closed
```

A closed ticket cannot be reopened or moved backward. Invalid transitions return `400`.

### Example requests

Register:
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"a@test.com","password":"password123"}'
```

Login:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"a@test.com","password":"password123"}'
```

Create ticket:
```bash
curl -X POST http://localhost:8080/tickets \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Bug in login","description":"Cannot login"}'
```

List tickets:
```bash
curl http://localhost:8080/tickets -H "Authorization: Bearer <TOKEN>"
```

Get ticket:
```bash
curl http://localhost:8080/tickets/<ID> -H "Authorization: Bearer <TOKEN>"
```

Update status:
```bash
curl -X PATCH http://localhost:8080/tickets/<ID>/status \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"status":"in_progress"}'
```

## Local Run (without Docker)

```bash
go mod tidy
go run ./cmd
curl http://localhost:8080/health
```

## Local Run (Docker) — required contract

```bash
docker build -t ticket-system .
docker run -p 8080:8080 ticket-system
curl http://localhost:8080/health
```

Expected health response:
```json
{"status": "ok"}
```

## Environment Variables

See `.env.example`.

| Variable    | Required | Default                   | Description                        |
|-------------|----------|----------------------------|--------------------------------------|
| JWT_SECRET  | Recommended | `dev-secret-change-me`  | Secret used to sign JWTs             |
| PORT        | No       | `8080`                     | Port the server listens on           |

To run with a custom secret via Docker:
```bash
docker run -p 8080:8080 -e JWT_SECRET=your-long-random-secret ticket-system
```

## Deployment

- Deployed URL: `<FILL IN AFTER DEPLOYING>`
- Public health check URL: `<FILL IN>/health`

(See "Deployment Steps" below for how to deploy for free on Render.)

## Assumptions

- No admin role, ticket assignment, or comments module, per assignment scope.
- Email is used as the unique login identifier; stored lower-cased/trimmed.
- Passwords must be at least 6 characters.
- Ticket ownership is enforced by returning `404` (not `403`) when a user requests a ticket that
  exists but isn't theirs, to avoid leaking existence of other users' tickets.
- In-memory storage means data resets on restart; acceptable per assignment scope ("no complex
  schema required... in-memory storage... is fine").
- `JWT_SECRET` defaults to a dev value if unset, so the app still runs out of the box, but a real
  secret must be set for anything beyond local testing (documented in `.env.example`).
