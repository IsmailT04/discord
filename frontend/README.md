# Frontend

React + TypeScript (Vite) SPA for the Discord-like web client.

## Architecture

Feature-based folders under `src/features/*`, with shared infrastructure in `src/shared/*` and app shell in `src/app/*`.

```
src/
  app/           → providers, router, layouts, global styles
  features/      → product features (auth, servers, chat, voice, …)
  shared/        → API client, WS client, UI kit, types, config
  assets/        → static images/fonts referenced by the app
```

Vite defaults (`main.tsx`, `App.tsx`, …) remain until features are wired; prefer moving routes into `app/router` as work starts.

## Recommended libraries (to add when coding starts)

- TanStack Query — server state
- Zustand (or RTK) — ephemeral UI / voice device state
- React Router — routing
- LiveKit client/components — voice & screen share
- OpenTelemetry Web SDK (optional RUM) → OTLP to SigNoz later

## Auth

- Cookie sessions (`credentials: 'include'`)
- CSRF header on mutating requests
- Same-origin behind nginx in production layout

## Remaining tasks

- [ ] npm install and lockfile commit
- [ ] Replace default Vite demo UI with app shell
- [ ] Wire providers (Query, auth session, WS)
- [ ] Implement features in shippable order (see root README plan)
- [ ] Virtualize long message lists
- [ ] Optional browser OTel traces to SigNoz

## Rules

- Features do not import other features’ internals deeply — use `shared` or explicit public exports
- No business secrets in frontend env; only public URLs (API, LiveKit, OTLP)
