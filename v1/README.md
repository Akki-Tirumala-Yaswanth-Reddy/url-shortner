# url-shortner v1

## Goal

v1 is the **minimal** working URL shortener implementation.
Focus:
- simple API
- simple redirect flow
- minimal dependencies

Non-goals (for v1):
- persistence
- analytics
- rate limiting
- multi-instance support

## How to run

### Prerequisites
- Go

### Run
```bash
cd v1
go run .
```

### Docker
```bash
cd v1
docker build -t url-shortner-v1 .
docker run -p 8080:8080 url-shortner-v1
```

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
{ "url": "localhost:8080/redirect/1" }
```

**Errors**

| Status | Reason |
|--------|--------|
| 400 | `user` or `url` is empty, or request body has unknown fields |

### GET /redirect/{id}
Redirects to the original URL.

**Response**
- `302 Found` — redirect to the original URL
- `400 Bad Request` — `id` is not a valid integer
- `404 Not Found` — `id` does not exist

## Example usage

```bash
# Shorten a URL
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"user": "Alice", "url": "https://example.com"}'
# → {"url":"localhost:8080/redirect/1"}

# Follow the redirect
curl -L http://localhost:8080/redirect/1
# → redirected to https://example.com
```

## Configuration

Port is hardcoded to **8080** (not configurable via environment variables).

## Implementation notes

- **Storage**: in-memory map (resets on restart)
- **ID generation**: A global counter is incremented on each `POST /create` and used as the short code (e.g. `1`, `2`, `3`).
- **Validation**: `user` and `url` fields must be non-empty. Unknown JSON fields in the request body are rejected.
- **Logging middleware**: Every request logs the HTTP method, URL path, and elapsed time.

## Known limitations

- Data is lost on restart (in-memory only)
- Counter resets to 0 on restart; short codes are re-used across restarts
- No URL format validation (any non-empty string is accepted as the destination URL)
- Not safe for multi-instance deployments
