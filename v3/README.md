# url-shortner v3

## Goal

v3 replaces the fragile in-memory counter from v2 with **DB-safe short-code generation** using PostgreSQL sequence IDs encoded to Base62.

Focus:
- PostgreSQL-backed storage (data survives restarts)
- Short codes derived from DB-generated `BIGSERIAL` id via Base62 encoding — no in-memory state
- Safe for restarts and multi-instance deployments
- Config via environment variables / `.env` file
- Structured packages (handlers, db, helpers, middleware, models)
- Logging middleware

Non-goals (for v3):
- Rate limiting
- Analytics
- TTL / link expiry

## How to run

### Prerequisites
- Go 1.25+
- PostgreSQL instance (or Docker for the Compose path)
- Redis instance (or Docker for the Compose path)

### Option 1 — run directly

From the `v3/` directory:

1. Start a PostgreSQL database and apply the schema:
   ```bash
   psql -U postgres -d url-shortner -f db/init.sql
   ```
2. Start a Redis instance (default port `6379`).
3. Set the required environment variables (edit `.env` or export them):
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/url-shortner?sslmode=disable"
   export REDIS_URL="redis://localhost:6379"
   ```
4. Run the server:
   ```bash
   go run .
   ```

### Option 2 — Docker Compose (recommended)

```bash
cd v3
docker compose -f Compose.yaml up
```

This starts the Go application, a PostgreSQL container, and a Redis container.
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

The `short_code` in the URL is the Base62-encoded database id (e.g. `1`, `B`, `1C`).

**Errors**

| Status | Reason |
|--------|--------|
| 400 | `user` or `url` is empty, or request body has unknown fields |
| 409 | Short code collision (should not happen under normal operation) |
| 500 | Database or server error |

### GET /redirect/{short_code}
Redirects to the original URL.

`short_code` is a Base62-encoded string (e.g. `1`, `B`, `1C`).

**Response**
- `302 Found` — redirect to the original URL
- `400 Bad Request` — short code is empty
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
| `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `PORT` | `8080` | Port the server listens on |

`DATABASE_URL` is read from the environment first; if unset, the server loads `.env` from the working directory.

Example `.env`:
```
DATABASE_URL="postgres://postgres:postgres@localhost:5432/url-shortner?sslmode=disable"
REDIS_URL="redis://localhost:6379"
```

## Implementation notes

- **Storage**: PostgreSQL via `pgx/v5` connection pool. Schema: `urls(id BIGSERIAL PK, user_name TEXT, short_code TEXT UNIQUE, original_url TEXT, created_at TIMESTAMPTZ)`.
- **Short-code generation**: On `POST /create` the handler inserts a row (without `short_code`), reads back the DB-generated `id` via `RETURNING id`, encodes it with Base62, and updates the row — all inside a single transaction.
- **Base62 encoding**: Uses digits `0-9`, uppercase `A-Z`, lowercase `a-z` (62 characters). Produces short, URL-safe codes that grow slowly (e.g. id 1 → `1`, id 62 → `10`, id 238328 → `1000`).
- **Redis caching**: On `GET /redirect/{short_code}`, the handler checks Redis first. On a cache hit the database is not queried. On a cache miss the database is queried and the result is written to Redis with a 24-hour TTL. New short codes are also written to Redis immediately after creation.
- **Validation**: `user` and `url` fields must be non-empty. Unknown JSON fields in the request body are rejected.
- **Logging middleware**: Every request logs the HTTP method, URL path, and elapsed time.

## Known limitations

- **No URL format validation**: Any non-empty string is accepted as the destination URL.
- **No TTL / link expiry**: Short URLs are permanent (the 24-hour Redis TTL only applies to the cache entry, not the underlying database record).
- **No rate limiting**: The API accepts unlimited requests per client.
- **No analytics**: Click counts or redirect metrics are not tracked.

## Improvements over v2

| v2 limitation | v3 fix |
|---------------|--------|
| In-memory counter resets on restart, causing `409 Conflict` collisions | Short codes derived from DB sequence — always unique across restarts |
| Not safe for multiple instances (counter is process-local) | All state lives in PostgreSQL; any number of instances can run concurrently |
| Short codes are plain incrementing integers | Short codes are Base62-encoded — shorter and less predictable |
