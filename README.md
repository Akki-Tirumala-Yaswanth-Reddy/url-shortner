# url-shortner (Go)

A URL shortener project written in Go.  
This repository contains **multiple versions** of the same project to showcase different implementations and tradeoffs (storage, scalability, APIs, reliability, etc.).

## Versions

| Version | Status | Storage | Highlights | Docs |
|--------:|:------:|:-------|:----------|:-----|
| v1 | ✅ Done | In-memory | Minimal working URL shortener | [v1 README](./v1/README.md) |
| v2 | ✅ Done | PostgreSQL | Persistent storage, Docker Compose, structured packages | [v2 README](./v2/README.md) |
| v3 | ✅ Done | PostgreSQL + Redis | Base62 short codes, Redis caching, restart-safe | [v3 README](./v3/README.md) |
| v4 | ✅ Done | PostgreSQL + Redis | Click analytics, health/readiness endpoints | [v4 README](./v4/README.md) |

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
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"user": "Alice", "url": "https://example.com"}'
```

Follow a redirect:
```bash
curl -L http://localhost:8080/redirect/1
```

## Repo structure

- `v1/` - First implementation (minimal, in-memory, focused on core workflow)
- `v2/` - Second implementation (PostgreSQL-backed persistence, Docker Compose)
- `v3/` - Third implementation (Base62 short codes, Redis caching, restart-safe)
- `v4/` - Fourth implementation (click analytics, health/readiness endpoints)

## Roadmap

- v5: rate limiting, TTL / link expiry
