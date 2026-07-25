# platform/http

Shared HTTP server utilities (not business routes).

## What lives here

- Router construction hooks
- JSON encode/decode helpers
- Standard success/error response shapes

## Remaining tasks

- [x] Problem+JSON or consistent `{ code, message }` errors
- [x] Request body size limits
- [x] Content-type enforcement helpers