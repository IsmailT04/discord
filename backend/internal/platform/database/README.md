# platform/database

Postgres connectivity and transaction helpers.

## What lives here

- Connection pool setup (pgx recommended)
- Ping/health for readiness
- Transaction helper used by application services

## Remaining tasks

- [ ] Pool config (max conns, lifetimes)
- [ ] OTel DB instrumentation
- [ ] Migrate runner is **not** here — see `cmd/migrate` + `/database`
