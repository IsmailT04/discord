# deploy/nginx

nginx as reverse proxy and load balancer in front of Go API instances.

## What lives here

- nginx.conf / site configs
- Upstream to API replicas
- SPA static serving (or proxy to frontend container)
- `/api` and `/ws` reverse proxy with WebSocket upgrade

## Remaining tasks

- [ ] Same-origin layout for cookie auth
- [ ] WebSocket `Upgrade` + long-lived timeouts
- [ ] TLS termination notes (local vs prod)
- [ ] Load-balance API without sticky sessions (shared Redis sessions)
- [ ] Do **not** proxy LiveKit media through this nginx unless intentionally designed (prefer direct LiveKit endpoint)
