# Platform

Cross-cutting technical capabilities. Not a business bounded context.

## What lives here

| Path | Responsibility |
|------|----------------|
| `config/` | Env loading, typed config structs |
| `database/` | Postgres pool, tx helpers |
| `redis/` | Redis client/pool |
| `http/` | Router setup, response helpers, problem+json |
| `middleware/` | Request ID, logging, auth, CSRF, CORS, rate limit, OTel HTTP |
| `auth/` | Shared cookie/CSRF helpers used by middleware + identity adapters |
| `observability/` | OpenTelemetry Tracer/Meter/Logger setup, SigNoz exporters, resource attrs |
| `storage/` | Shared blob port types if needed outside Chat |
| `errors/` | Mapped application/domain → HTTP errors |
| `clock/` | Clock port for testability |

## Remaining tasks

- [ ] Config with validation (fail fast on missing secrets)
- [ ] Postgres + Redis constructors with health checks
- [ ] OTel: traces, metrics, log correlation; export OTLP to SigNoz
- [ ] HTTP middleware stack order documented
- [ ] Consistent error envelope
- [ ] Never log secrets, tokens, or raw passwords
