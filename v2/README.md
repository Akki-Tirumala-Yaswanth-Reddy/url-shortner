# url-shortner v2

## Goal

v2 adds **persistent storage** via PostgreSQL and a more structured codebase.
Focus:
- PostgreSQL-backed storage (data survives restarts)
- Config via environment variables / `.env` file
- Structured packages (handlers, db, helpers, middleware, models)
- Logging middleware
- Docker and Docker Compose support

Non-goals (for v2):
- Rate limiting
- Analytics
- Multi-instance support
- TTL / link expiry

## How to run

### Prerequisites
- Go
- PostgreSQL instance (or Docker for the Compose path)

### Option 1 — run directly

1. Start a PostgreSQL database and apply the schema:
   ```bash
   psql -U postgres -d url-shortner -f db/init.sql
   ```
2. Set the database URL (edit `.env` or export the variable):
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/url-shortner?sslmode=disable"
   ```
3. Run the server:
   ```bash
   cd v2
   go run .
   ```

### Option 2 — Docker Compose (recommended)

```bash
cd v2
docker compose -f Compose.yaml up
```

This starts both the Go application and a PostgreSQL container.
The database schema is applied automatically via `db/init.sql`.

## API

### POST /create
Creates a short URL.

**Request**
```json
{
  "user": "Alice",
  "url": "https://example.com"
}
```

**Response**
```json
{ "url": "localhost:8080/redirect/1", "id": 1 }
```

**Errors**

| Status | Reason |
|--------|--------|
| 400 | `user` or `url` is empty, or request body has unknown fields |
| 409 | Short code already exists in the database |
| 500 | Database or server error |

### GET /redirect/{short_code}
Redirects to the original URL.

`short_code` must be a numeric string (e.g. `1`, `42`).

**Response**
- `302 Found` — redirect to the original URL
- `400 Bad Request` — short code is not a valid number
- `404 Not Found` — short code does not exist in the database

## Example usage

```bash
# Shorten a URL
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"user": "Alice", "url": "https://example.com"}'
# → {"url":"localhost:8080/redirect/1","id":1}

# Follow the redirect
curl -L http://localhost:8080/redirect/1
# → redirected to https://example.com
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | *(required)* | PostgreSQL connection string |
| `PORT` | `8080` | Port the server listens on |

`DATABASE_URL` is read from the environment first; if unset, the server loads `.env` from the working directory.

Example `.env`:
```
DATABASE_URL="postgres://postgres:postgres@localhost:5432/url-shortner?sslmode=disable"
```

## Implementation notes

- **Storage**: PostgreSQL via `pgx/v5` connection pool. Schema: `urls(id SERIAL PK, user_name TEXT, short_code TEXT UNIQUE, original_url TEXT, created_at TIMESTAMPTZ)`.
- **ID generation**: An in-memory counter (protected by a `sync.RWMutex`) is incremented on each `POST /create` and used as the short code.
- **Validation**: `user` and `url` fields must be non-empty. Unknown JSON fields in the request body are rejected. The redirect short code must parse as an integer.
- **Logging middleware**: Every request logs the HTTP method, URL path, and elapsed time.

## Known limitations

- **Counter resets on restart**: The short-code counter is in-memory and starts from 0 on every restart. If the database already contains entries, new short codes will collide with existing ones, returning `409 Conflict` until the counter surpasses the highest stored short code.
- **Not safe for multi-instance deployments**: The counter is local to each process; running multiple instances will produce duplicate short codes.
- **No URL format validation**: Any non-empty string is accepted as the destination URL.
