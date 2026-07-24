# deploy/compose

Docker Compose definitions for local / single-host runs.

## What lives here

- `docker-compose.yml` (and overrides) for: API, Postgres, Redis, OTel collector, SigNoz (or SigNoz lite), references to LiveKit
- Env sample files

## Remaining tasks

- [ ] Base compose with Postgres + Redis + API
- [ ] Optional profile for SigNoz / OTel collector
- [ ] Optional profile linking LiveKit using config from `/livekit`
- [ ] Document ports and volumes
