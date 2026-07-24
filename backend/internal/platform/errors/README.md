# platform/errors

Error mapping between layers.

## What lives here

- Sentinel / typed errors bridging domain → HTTP status
- No framework leakage into domain

## Remaining tasks

- [ ] Map not-found, forbidden, conflict, validation, unauthorized
- [ ] Ensure WS and HTTP share the same code vocabulary where useful
