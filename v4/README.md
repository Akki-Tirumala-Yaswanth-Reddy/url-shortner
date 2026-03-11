# url-shortner v4

## Goal

v4 builds on v3 by adding **per-link click analytics** and **operational endpoints** (health and readiness checks).

New in v4:
- **Click tracking** — every redirect atomically increments `click_count` and sets `last_accessed_at` in PostgreSQL (no race conditions, survives restarts)
- **Stats API** — `GET /stats/{short_code}` returns per-link analytics
- **Health endpoints** — `GET /healthCheck` (liveness) and `GET /readyCheck` (DB readiness)

Non-goals (for v4):
- Authentication / authorization
- UI dashboard
- Per-day / time-series analytics

## How to run

### Prerequisites
- Go 1.25+
- Docker & Docker Compose (for the Compose path)

### Run with Docker Compose (recommended)

```bash
cd v4
docker compose up --build
```

This starts PostgreSQL, Redis, and the app. The schema is applied automatically via `db/init.sql`.

### Run directly

1. Start PostgreSQL and Redis, then apply the schema:
   ```bash
   psql -U postgres -d url-shortner -f db/init.sql
   ```
2. Export env vars:
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/url-shortner?sslmode=disable"
   export REDIS_URL="redis://localhost:6379"
   ```
3. Run:
   ```bash
   go run .
   ```

## API

### POST /create
Creates a short URL.

**Request**
```json
{ "user": "Alice", "url": "https://example.com" }
```

**Response** `200`
```json
{ "url": "localhost:8080/redirect/1", "id": 1 }
```

**Errors**

| Status | Reason |
|--------|--------|
| 400 | `user` or `url` is empty, or request body has unknown fields |
| 409 | Short code collision (should not happen under normal operation) |
| 500 | Database or server error |

### GET /redirect/{short_code}
Redirects to the original URL. Atomically increments `click_count` and updates `last_accessed_at`.

- `302 Found` — redirect to the original URL
- `400 Bad Request` — short code is empty
- `404 Not Found` — short code does not exist

### GET /stats/{short_code}
Returns per-link analytics.

**Response** `200`
```json
{
  "short_code": "1",
  "original_url": "https://example.com",
  "created_at": "2026-03-10T00:00:00Z",
  "click_count": 3,
  "last_accessed_at": "2026-03-10T12:34:56Z"
}
```

**Errors**

| Status | Reason |
|--------|--------|
| 404 | Short code does not exist |
| 500 | Database or server error |

### GET /healthCheck
Always returns `200 OK` with body `ok`.

### GET /readyCheck
Pings the database. Returns `200 OK` with body `ok` if reachable, `503 Service Unavailable` with body `db not ready` otherwise.

## Example walkthrough

```bash
# 1. Create a short URL
curl -s -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"user": "Alice", "url": "https://example.com"}'
# → {"url":"localhost:8080/redirect/1","id":1}

# 2. Hit redirect a few times to generate clicks
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/redirect/1
# → 302
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/redirect/1
# → 302
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/redirect/1
# → 302

# 3. Fetch stats — click_count should be 3
curl -s http://localhost:8080/stats/1 | python3 -m json.tool
# {
#     "short_code": "1",
#     "original_url": "https://example.com",
#     "created_at": "2026-03-10T...",
#     "click_count": 3,
#     "last_accessed_at": "2026-03-10T..."
# }

# 4. Health checks
curl -s http://localhost:8080/healthCheck   # → ok
curl -s http://localhost:8080/readyCheck    # → ok
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | *(required)* | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `PORT` | `8080` | Port the server listens on |

`DATABASE_URL` is read from the environment first; if unset, the server loads `.env` from the working directory.

Example `.env`:
```
DATABASE_URL="postgres://postgres:postgres@localhost:5432/url-shortner?sslmode=disable"
REDIS_URL="redis://localhost:6379"
```

## Implementation notes

- **Storage**: PostgreSQL via `pgx/v5` connection pool. Schema: `urls(id BIGSERIAL PK, user_name TEXT, short_code TEXT UNIQUE, original_url TEXT, created_at TIMESTAMPTZ, click_count BIGINT DEFAULT 0, last_accessed_at TIMESTAMPTZ NULL)`.
- **Short-code generation**: On `POST /create` the handler inserts a row (without `short_code`), reads back the DB-generated `id` via `RETURNING id`, encodes it with Base62, and updates the row — all inside a single transaction.
- **Base62 encoding**: Uses digits `0-9`, uppercase `A-Z`, lowercase `a-z` (62 characters). Produces short, URL-safe codes that grow slowly (e.g. id 1 → `1`, id 62 → `10`, id 238328 → `1000`).
- **Click tracking**: On `GET /redirect/{short_code}`, `click_count` is atomically incremented and `last_accessed_at` is set to `now()` in a single `UPDATE ... RETURNING` query.
- **Redis caching**: On `GET /redirect/{short_code}`, the handler checks Redis first. On a cache hit the database lookup is skipped but the click metrics are still updated. On a cache miss the database is queried and the result is written to Redis with a 24-hour TTL. New short codes are also written to Redis immediately after creation.
- **Validation**: `user` and `url` fields must be non-empty. Unknown JSON fields in the request body are rejected.
- **Logging middleware**: Every request (except `/healthCheck` and `/readyCheck`) logs the HTTP method, URL path, and elapsed time.

## Known limitations

- **No URL format validation**: Any non-empty string is accepted as the destination URL.
- **No TTL / link expiry**: Short URLs are permanent (the 24-hour Redis TTL only applies to the cache entry, not the underlying database record).
- **No rate limiting**: The API accepts unlimited requests per client.
- **No per-day analytics**: Click counts are cumulative totals only; time-series data is not stored.

## Improvements over v3

| v3 limitation | v4 fix |
|---------------|--------|
| No click tracking | Every redirect atomically increments `click_count` and records `last_accessed_at` in the database |
| No way to inspect link usage | New `GET /stats/{short_code}` endpoint returns per-link analytics |
| No operational probes | New `GET /healthCheck` (liveness) and `GET /readyCheck` (DB readiness) endpoints |
