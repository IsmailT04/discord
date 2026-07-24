# database

PostgreSQL schema migrations and optional seed data.

## Layout

```
database/
  migrations/   → ordered SQL migration files
  seeds/        → development-only seed SQL (never required in prod)
```

## Conventions

- Migrations are the **source of truth** for schema
- Naming: `000001_init.up.sql` / `000001_init.down.sql` (or tool-native format)
- Applied by `backend/cmd/migrate`
- Attachments stored as `BYTEA` (raw bytes), **not** base64 text
- Prefer UUID/ULID primary keys consistently

## Suggested migration sequence (high level)

1. Users / sessions (if DB-backed) / profiles  
2. Servers, members, channels  
3. Roles, permissions, overwrites  
4. Messages, reactions, attachments  
5. Friendships, DM conversations  
6. Notifications + preferences  
7. Audit log / pins / extras  

## Remaining tasks

- [ ] Choose migration tool and file naming
- [ ] Write migrations per shippable update (do not big-bang all tables)
- [ ] Add indexes for message channel+created_at, membership lookups
- [ ] Document BYTEA size limits enforced in app
- [ ] Seeds for local demo users/servers (optional)

## Rules

- No application code here — SQL only
- Do not commit production dumps or secrets
