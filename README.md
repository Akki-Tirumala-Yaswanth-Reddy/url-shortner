# url-shortner (Go)

A URL shortener project written in Go.  
This repository contains **multiple versions** of the same project to showcase different implementations and tradeoffs (storage, scalability, APIs, reliability, etc.).

## Versions

| Version | Status | Storage | Highlights | Docs |
|--------:|:------:|:-------|:----------|:-----|
| v1 | ✅ Done | In-memory (example) | Minimal working URL shortener | [v1 README](./v1/README.md) |
| v2 | 🟡 Planned | TBD | Persistence + better validation | [v2 README](./v2/README.md) |

> Tip: Each version is intended to be **self-contained**. Read the version README for run steps and API.

## Quickstart (v1)

### Prerequisites
- Go (recommended: latest stable)

### Run
```bash
cd v1
go run .
```

### Example usage
Shorten a URL:
```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"user":"user's name", "url":"https://example.com"}'
```

Redirect:
```bash
curl -I http://localhost:8080/<short-code>
```

## Repo structure

- `v1/` - First implementation (minimal, focused on core workflow)
- `v2/` - Next iteration (planned)

## Roadmap

- v2: persistent storage, config via env, structured logging, tests
- v3: rate limiting, analytics, caching, expiry/TTL, docker compose, etc.
