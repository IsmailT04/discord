# migrations

Ordered PostgreSQL migration scripts.

## What lives here

- `*.up.sql` / `*.down.sql` (or single-folder versioned files per migrator)
- One logical change per migration when practical

## Remaining tasks

- [ ] Initial users/identity tables
- [ ] Directory tables (servers, channels, roles, …)
- [ ] Chat tables (messages, reactions, attachments BYTEA)
- [ ] Social tables (friends, conversations)
- [ ] Notifications tables
- [ ] Keep downs usable for local dev; be careful in shared envs
