# platform/storage

Shared storage port types (optional).

## What lives here

- Generic blob interfaces if Chat and Profile avatars share one port
- Postgres BYTEA adapter may live in Chat; profile avatars can reuse this port later

## Remaining tasks

- [ ] Decide single BlobStore port vs context-local stores
- [ ] Size/MIME validation helpers
