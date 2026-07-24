# identity/adapters/session

## What lives here

Redis/Postgres session and refresh-token store.

## Remaining tasks

- [ ] Implement port interfaces for this adapter
- [ ] Map to/from domain types at the boundary
- [ ] Add OTel spans for external I/O
- [ ] Handle adapter-specific errors → platform/domain errors
