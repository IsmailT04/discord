# cmd/migrate

CLI entrypoint to apply database migrations.

## What lives here

- Migration runner that reads SQL files from repo-root `database/migrations`
- Version tracking against Postgres

## Remaining tasks

- [ ] Choose and wire migrator (e.g. golang-migrate, goose, atlas)
- [ ] Commands: `up`, `down`, `status`, `force` (ops-safe)
- [ ] Fail CI if migrations are not sequential / checksum-valid
- [ ] Document how seeds in `database/seeds` are applied (dev only)
