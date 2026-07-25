# platform/logger

Process and request-scoped logging built on zap.

## What lives here

- `New(cfg, extraCores...)` — stdout logger (console in dev/staging, JSON in production)
- Prefer `observability.Providers.Logger()` after OTel init (tees otelzap core)
- Context helpers: `ToContext`, `FromContext`, `WithContext`
- `WithTrace(ctx)` — adds `trace_id` / `span_id` from the active OTel span
- Runtime `SetLevel` via atomic level
- `Sync` for graceful shutdown

## Remaining tasks

- [x] Zap logger with env-aware encoding
- [x] Context propagation helpers
- [x] Dynamic level change
- [x] Bound to observability via otelzap (from Providers)
- [ ] Optional `LOG_LEVEL` env in config
- [ ] Middleware that injects `request_id` + `WithTrace` into context logger
