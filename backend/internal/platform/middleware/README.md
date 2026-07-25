# platform/middleware

HTTP middleware chain.

## What lives here

- Request ID (context + `X-Request-ID` response echo)
- Recover / panic handler (zap + `httpx` JSON 500)
- Access logging (correlated with OTel) — `StatusResponseWriter` captures status
- CSRF validation (double-submit cookie; safe methods bypass)
- Rate limiting (in-memory token bucket; key by IP or custom)
- OTel HTTP middleware (`SpanTracker`)
- Session auth — `LoadSession` (optional) + `RequireAuth` (401 if missing)
- CORS — deferred (prefer same-origin via nginx)

## Suggested order (outermost first)

```text
Recovery → RequestID → SpanTracker → AccessLog → RateLimit → LoadSession → CSRF → handler
```

`Chain` applies so the **first** listed middleware is outermost.

Public / optional-auth group (wired in `cmd/api`):

```go
middleware.Chain(mux,
    middleware.Recovery,
    middleware.RequestID,
    middleware.SpanTracker,
    middleware.AccessLog,
    middleware.RateLimit(limiter),
    middleware.LoadSession(sessions),
    middleware.CSRF,
)
```

Authenticated group (add when Identity routes exist):

```go
middleware.Chain(protectedMux,
    middleware.Recovery,
    middleware.RequestID,
    middleware.SpanTracker,
    middleware.AccessLog,
    middleware.RateLimit(limiter),
    middleware.LoadSession(sessions),
    middleware.RequireAuth,
    middleware.CSRF,
)
```

## Remaining tasks

- [x] Authenticated vs public route groups (`LoadSession` + `RequireAuth`)
- [x] Wire stack in `cmd/api` (healthz/readyz; auth routes are yours)
- [x] Recovery uses zap + httpx
- [x] Echo `X-Request-ID` on responses
- [ ] Optional Redis-backed rate limiter for multi-instance
- [ ] CORS only if cross-origin SPA is required
