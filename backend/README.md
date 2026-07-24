# Backend

Go API for the Discord-like web application.

## Architecture

Hexagonal architecture (ports & adapters) with DDD-style bounded contexts under `internal/`.

```
cmd/           → process entrypoints (API, migrate)
internal/      → bounded contexts + platform (not importable by others as a library)
pkg/           → intentionally small shared libraries safe for external import
api/openapi/   → OpenAPI contracts
deploy/        → compose, nginx, k8s, observability (SigNoz/OTel collector)
scripts/       → operational scripts (no business logic)
```

## Bounded contexts

| Context | Responsibility |
|---------|----------------|
| `identity` | Users, profiles, sessions, CSRF, auth cookies |
| `directory` | Servers, channels, roles, permissions, invites, membership |
| `chat` | Messages, replies, reactions, attachments (BYTEA) |
| `social` | Friends, DMs, group DMs |
| `realtime` | WebSocket hub, presence, typing, event fanout |
| `media` | LiveKit token issuance and voice-room policy |
| `notification` | In-app notifications and preferences |
| `platform` | Config, DB, Redis, HTTP, middleware, OTel, storage ports |
| `sharedkernel` | Cross-context primitives (IDs, pagination cursors) — keep tiny |

## Layering (per context)

- `domain/` — entities, value objects, domain errors, permission rules
- `application/` — use cases / application services
- `ports/` — interfaces (repositories, publishers, token issuers)
- `adapters/` — HTTP, persistence, Redis, LiveKit, etc.

## Observability

OpenTelemetry instrumentation lives in `internal/platform/observability` and middleware.
Traces, metrics, and logs export to SigNoz (see `deploy/observability`).

## Remaining tasks

- [ ] Wire `cmd/api` composition root (dependency injection of adapters → use cases)
- [ ] Wire `cmd/migrate` to apply SQL from `/database/migrations`
- [ ] Implement contexts in shippable order (see root `README.md` plan)
- [ ] Add OpenAPI specs under `api/openapi`
- [ ] Add Compose/K8s/nginx/SigNoz deploy manifests under `deploy/`
- [ ] CI: lint, test, migrate-check, build

## Rules

- Domain must not import adapters or framework packages
- Application depends on ports only
- Adapters implement ports
- Cross-context calls go through application services or explicit anti-corruption interfaces — no deep domain coupling
