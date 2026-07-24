# cmd/api

Process entrypoint for the HTTP + WebSocket API server.

## What lives here

- `main.go` — composition root only: load config, build adapters, mount routes, start server, graceful shutdown
- No business logic

## Remaining tasks

- [ ] Load config from env
- [ ] Init Postgres, Redis, OTel (tracer/meter/logger providers → SigNoz)
- [ ] Construct repositories, use cases, HTTP/WS handlers per context
- [ ] Register routes and middleware (auth, CSRF, request ID, metrics, recovery)
- [ ] Health/ready probes (`/healthz`, `/readyz`)
- [ ] Graceful shutdown with drain timeout
