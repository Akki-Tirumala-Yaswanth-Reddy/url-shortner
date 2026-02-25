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
go run .
```
## API

### POST /create
Creates a short URL.

**Request**
```json
{ "url": "https://example.com" }
```

**Response**
```json
{ "short_url": "http://localhost:8080/123" }
```

### GET /redirect/{extra}
Redirects to the original URL.

## Configuration

- `PORT` (default: 8080)

## Implementation notes

- Storage: in-memory map (resets on restart)
- ID generation: A counter is used to keep of track of 
- Validation: (what you validate, what you don’t)

## Known limitations

- Data lost on restart
- No collision handling beyond X
- Not safe for multi-instance deployments