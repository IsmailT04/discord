# deploy/compose

Docker Compose for local infra. The Go API runs on the host via `make dev-api`.

## Files

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Postgres + Redis; optional `observability` profile |
| `otel-collector-config.yaml` | OTLP → Jaeger (+ debug) |
| `.env.example` | Compose credentials / ports |
| `.env` | Local compose env (gitignored; copy from example) |

## Ports

| Service | Host port |
|---------|-----------|
| Postgres | 5432 |
| Redis | 6379 |
| OTLP HTTP (profile) | 4318 |
| OTLP gRPC (profile) | 4317 |
| Jaeger UI (profile) | 16686 |

## Makefile

```bash
make up          # postgres + redis (waits for healthy)
make down
make logs
make signoz-up   # + otel-collector + jaeger
make signoz-down
make dev-api     # host API against localhost DB/Redis
```

Credentials must match `backend/.env` (`DB_*`, `REDIS_*`).

## Remaining tasks

- [x] Base compose with Postgres + Redis
- [x] Optional observability profile (collector + Jaeger UI)
- [ ] Full SigNoz stack (replace Jaeger when you want production-like dashboards)
- [ ] Optional LiveKit service profile
- [x] Document ports and volumes
