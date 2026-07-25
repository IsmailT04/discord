# cmd/api

Process entrypoint for the HTTP API server.

## What lives here

- `main.go` — composition root: config, OTel, logger, DB, Redis, middleware, health routes, graceful shutdown
- No Identity business logic (register/login/etc. — you wire those)

## Current routes

| Method | Path | Auth |
|--------|------|------|
| GET | `/healthz` | public |
| GET | `/readyz` | public (checks Postgres + Redis) |

Middleware stack: Recovery → RequestID → SpanTracker → AccessLog → RateLimit → LoadSession → CSRF

## Remaining tasks

- [x] Load config from env (`.env` via godotenv)
- [x] Init Postgres, Redis, OTel
- [x] Register middleware + health/ready
- [x] Graceful shutdown
- [ ] Mount Identity HTTP adapters when ready
- [ ] Split public vs `RequireAuth` route groups for protected APIs
