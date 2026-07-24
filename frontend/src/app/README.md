# src/app

Application shell: providers, routing, layouts, global styles.

## What lives here

| Path | Contents |
|------|----------|
| `providers/` | QueryClient, auth session, theme, WS provider |
| `router/` | Route table, protected/public route guards |
| `layouts/` | App chrome: server rail, channel sidebar, main pane, voice dock |
| `styles/` | Global CSS variables, reset, typography tokens |

## Remaining tasks

- [ ] Auth-aware router (login vs app shell)
- [ ] Layout matching Discord-like 3-pane structure
- [ ] Provider composition order documented
- [ ] Global error / toast boundary
