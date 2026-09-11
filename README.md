# Ticket System

A small backend service for a ticket system where a user can register, log in, create tickets, view only their own tickets, and update the status of their own tickets.

## Live Deployment

- **Deployed Application URL:** https://ticket-system-8kgq.onrender.com
- **Public Health Check URL:** https://ticket-system-8kgq.onrender.com/health

> Note: this is hosted on Render's free tier, so the service may spin down after periods of inactivity. The first request after idle can take 20–50 seconds to respond while it wakes up.

## Tech Stack

- **Language:** Go
- **Auth:** JWT (github.com/golang-jwt/jwt/v5)
- **Password hashing:** bcrypt (golang.org/x/crypto/bcrypt)
- **Storage:** In-memory store (no external database required)
- **Router:** Go 1.22+ standard library `net/http` `ServeMux` (method + path pattern routing)

## Project Structure

```
.
├── cmd/
│   └── main.go                  # entrypoint, route wiring
├── internal/
│   ├── auth/                    # JWT issuing/parsing, password hashing
│   ├── handlers/                # HTTP handlers (auth, tickets)
│   │   └── respond/             # JSON response helpers
│   ├── middleware/               # JWT auth middleware
│   ├── models/                   # Ticket/User types, status transitions
│   └── store/                    # in-memory data store
├── Dockerfile
├── go.mod / go.sum
├── .env.example
└── README.md
```

## Local Run (without Docker)

```bash
go run ./cmd
```

The server listens on port `8080` by default (or `$PORT` if set). Requires `JWT_SECRET` to be set in the environment — see `.env.example`.

## Docker Run Contract

```bash
docker build -t ticket-system .
docker run -p 8080:8080 --env-file .env ticket-system
curl http://localhost:8080/health
```

Expected health response:

```json
{"status": "ok"}
```

## Environment Variables

See `.env.example`:

```
JWT_SECRET=your-secret-key-here
PORT=8080
```

- `JWT_SECRET` — required. Used to sign and verify JWTs. The app will fail to start if this is not set.
- `PORT` — optional, defaults to `8080`.

## API Reference

All responses are JSON. Protected endpoints require an `Authorization: Bearer <token>` header, obtained from `/auth/register` or `/auth/login`.

Base URL (deployed): `https://ticket-system-8kgq.onrender.com`
Base URL (local): `http://localhost:8080`

### Health Check

**GET /health**
Deployed: https://ticket-system-8kgq.onrender.com/health
Local: http://localhost:8080/health

```bash
curl https://ticket-system-8kgq.onrender.com/health
```

Response `200 OK`:
```json
{"status": "ok"}
```

---

### Register

**POST /auth/register**
Deployed: https://ticket-system-8kgq.onrender.com/auth/register
Local: http://localhost:8080/auth/register

```bash
curl -X POST https://ticket-system-8kgq.onrender.com/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
```

Response `201 Created`:
```json
{
  "token": "<jwt>",
  "user": { "id": "usr_00000001", "email": "test@example.com" }
}
```

Validation:
- Email must be non-empty and contain `@`.
- Password must be at least 6 characters.
- Duplicate email returns `409 Conflict`.

---

### Login

**POST /auth/login**
Deployed: https://ticket-system-8kgq.onrender.com/auth/login
Local: http://localhost:8080/auth/login

```bash
curl -X POST https://ticket-system-8kgq.onrender.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
```

Response `200 OK`:
```json
{
  "token": "<jwt>",
  "user": { "id": "usr_00000001", "email": "test@example.com" }
}
```

Invalid credentials return `401 Unauthorized`.

---

### Create Ticket

**POST /tickets** (protected)
Deployed: https://ticket-system-8kgq.onrender.com/tickets
Local: http://localhost:8080/tickets

```bash
curl -X POST https://ticket-system-8kgq.onrender.com/tickets \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Printer is broken","description":"No toner"}'
```

Response `201 Created`:
```json
{
  "id": "tkt_00000001",
  "user_id": "usr_00000001",
  "title": "Printer is broken",
  "description": "No toner",
  "status": "open",
  "created_at": "2026-09-11T15:41:40Z",
  "updated_at": "2026-09-11T15:41:40Z"
}
```

Validation: `title` is required and cannot be blank.

---

### List My Tickets

**GET /tickets** (protected)
Deployed: https://ticket-system-8kgq.onrender.com/tickets
Local: http://localhost:8080/tickets

```bash
curl https://ticket-system-8kgq.onrender.com/tickets \
  -H "Authorization: Bearer <token>"
```

Response `200 OK`: array of the logged-in user's tickets only.

---

### Get Ticket by ID

**GET /tickets/{id}** (protected)
Deployed: https://ticket-system-8kgq.onrender.com/tickets/{id}
Local: http://localhost:8080/tickets/{id}

```bash
curl https://ticket-system-8kgq.onrender.com/tickets/tkt_00000001 \
  -H "Authorization: Bearer <token>"
```

Response `200 OK`: the ticket, if it belongs to the caller.
Response `404 Not Found`: if the ticket doesn't exist, or belongs to another user (existence is not leaked).

---

### Update Ticket Status

**PATCH /tickets/{id}/status** (protected)
Deployed: https://ticket-system-8kgq.onrender.com/tickets/{id}/status
Local: http://localhost:8080/tickets/{id}/status

```bash
curl -X PATCH https://ticket-system-8kgq.onrender.com/tickets/tkt_00000001/status \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"in_progress"}'
```

Response `200 OK`: the updated ticket.

Status flow rules:
- Allowed statuses: `open`, `in_progress`, `closed`.
- Valid transitions: `open → in_progress → closed`.
- A `closed` ticket cannot be moved back to `open` or `in_progress` — returns `409 Conflict`.
- Setting a ticket to its current status returns `400 Bad Request`.
- Invalid status values return `400 Bad Request`.
- Updating another user's ticket, or a nonexistent ticket, returns `404 Not Found`.

---

## Status Codes Summary

| Code | Meaning |
|------|---------|
| 200  | Success (read/update) |
| 201  | Resource created (register, create ticket) |
| 400  | Invalid request body / invalid status |
| 401  | Missing/invalid credentials or token |
| 404  | Resource not found or not owned by caller |
| 409  | Conflict (duplicate email, invalid status transition, already-set status) |
| 500  | Internal server error |

## Assumptions

- No admin role, ticket assignment, or comments module — out of scope per the assignment.
- Storage is in-memory; data resets on service restart (acceptable per assignment scope, which allows in-memory storage).
- Ticket and user IDs are simple sequential identifiers (`usr_00000001`, `tkt_00000001`) generated by the in-memory store.
- Accessing another user's ticket returns `404` rather than `403`, to avoid leaking the existence of tickets that don't belong to the caller.
- `JWT_SECRET` is required at startup; the app will not start without it (no insecure fallback in production).
- No frontend is included — this assignment specifies a backend service only.

## Deployment Notes

- Deployed on [Render](https://render.com) (free tier) using the included `Dockerfile`.
- Root Directory left blank in Render settings, since `go.mod`, `cmd/`, and `internal/` all live at the repository root.
- Render auto-deploys on every push to the `main` branch.