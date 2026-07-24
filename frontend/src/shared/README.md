# shared

Cross-feature frontend infrastructure and design system primitives.

## What lives here

| Path | Contents |
|------|----------|
| `api/` | fetch/axios client, CSRF injection, error parsing |
| `ws/` | WebSocket client, reconnect, event router |
| `ui/` | Reusable presentational components (buttons, inputs, modals) |
| `lib/` | Pure helpers (dates, cn/classnames, validators) |
| `types/` | Shared DTO/types aligned with API |
| `config/` | Public env (API base, WS URL, LiveKit URL) |
| `hooks/` | Generic hooks (not feature-specific) |

## Remaining tasks

- [ ] API client with credentials + CSRF
- [ ] WS client with auth cookie and backoff reconnect
- [ ] Minimal UI kit sufficient for auth + chat
- [ ] Typed event map for realtime events
- [ ] Config via `import.meta.env`
