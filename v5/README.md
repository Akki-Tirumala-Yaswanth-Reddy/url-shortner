# url-shortner v5

## Goal

v5 builds on v4 by adding **batched analytics flushes**, **Redis-backed rate limiting**, and **automatic URL expiration**.

New in v5:
- **In-memory analytics buffering** - redirects increment per-link counters in memory, then a background goroutine flushes the counts to PostgreSQL every 1 minute
- **Stats API** - `GET /stats/{short_code}` still returns per-link analytics from PostgreSQL
- **Rate limiting** - Redis is used to enforce per-route request limits
- **URL expiration** - URLs older than 2 months are deleted by a background goroutine every 12 hours
- **Health endpoints** - `GET /healthCheck` (liveness) and `GET /readyCheck` (DB readiness)

Non-goals (for v5):
- Authentication / authorization
- UI dashboard
- Per-day / time-series analytics

## How to run

### Prerequisites
- Go 1.25+
- Docker and Docker Compose (for the Compose path)

### Run with Docker Compose (recommended)

```bash
cd v5
docker compose -f Compose.yaml up
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

### GET /redirect/{short_code}
Redirects to the original URL. The redirect is rate limited and the click count is buffered in memory before being flushed to PostgreSQL.

- `302 Found` - redirect
- `404 Not Found` - short code does not exist
- `429 Too Many Requests` - rate limit exceeded

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

- `404 Not Found` - short code does not exist
- `500 Internal Server Error` - DB error

### GET /healthCheck
Always returns `200 OK`.

### GET /readyCheck
Pings the database. Returns `200 OK` if reachable, `503 Service Unavailable` otherwise.

## Example walkthrough

```bash
# 1. Create a short URL
curl -s -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"user": "Alice", "url": "https://example.com"}'
# -> {"url":"localhost:8080/redirect/1","id":1}

# 2. Hit redirect a few times to generate clicks
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/redirect/1
# -> 302
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/redirect/1
# -> 302
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/redirect/1
# -> 302

# 3. Wait for the analytics flush job to persist the in-memory counts

# 4. Fetch stats - click_count should be 3 after the flush runs
curl -s http://localhost:8080/stats/1 | python3 -m json.tool
# {
#     "short_code": "1",
#     "original_url": "https://example.com",
#     "created_at": "2026-03-10T...",
#     "click_count": 3,
#     "last_accessed_at": "2026-03-10T..."
# }

# 5. Health checks
curl -s http://localhost:8080/healthCheck   # -> ok
curl -s http://localhost:8080/readyCheck    # -> ok
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | *(required)* | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `EXPIRATION_LIMIT` | `2 months` | Age threshold for deleting URLs |
| `PORT` | `8080` | Port the server listens on |

`DATABASE_URL` is read from the environment first; if unset, the server loads `.env` from the working directory.

Example `.env`:
```
DATABASE_URL="postgres://postgres:postgres@localhost:5432/url-shortner?sslmode=disable"
REDIS_URL="redis://localhost:6379"
EXPIRATION_LIMIT="2 months"
```

## Implementation notes

- **Storage**: PostgreSQL via `pgx/v5` connection pool. Schema: `urls(id BIGSERIAL PK, user_name TEXT, short_code TEXT UNIQUE, original_url TEXT, created_at TIMESTAMPTZ, click_count BIGINT DEFAULT 0, last_accessed_at TIMESTAMPTZ NULL)`.
- **Short-code generation**: On `POST /create` the handler inserts a row (without `short_code`), reads back the DB-generated `id` via `RETURNING id`, encodes it with Base62, and updates the row - all inside a single transaction.
- **Base62 encoding**: Uses digits `0-9`, uppercase `A-Z`, lowercase `a-z` (62 characters). Produces short, URL-safe codes that grow slowly (e.g. id 1 -> `1`, id 62 -> `10`, id 238328 -> `1000`).
- **Analytics buffering**: Redirects increment an in-memory map protected by a mutex, and a background goroutine flushes the accumulated counts to PostgreSQL every minute.
- **Rate limiting**: Redis stores request counters for rate limiting, which helps protect high-traffic routes from abuse.
- **Expiration cleanup**: A background goroutine runs every 12 hours and deletes rows whose `created_at` is older than the configured expiration window.
- **Validation**: `user` and `url` fields must be non-empty. Unknown JSON fields in the request body are rejected.
- **Logging middleware**: Every request logs the HTTP method, URL path, and elapsed time.

## Improvements over v4

| v4 limitation | v5 fix |
|---------------|--------|
| Every redirect wrote analytics immediately to PostgreSQL | Clicks are buffered in memory and flushed once per minute to reduce write pressure |
| No request throttling | Redis-backed rate limiting protects the API from abuse |
| Old URLs stayed in the database forever | URLs older than 2 months are cleaned up automatically every 12 hours |
| No operational endpoints | `GET /healthCheck` (liveness) and `GET /readyCheck` (DB readiness) enable container health probes |
