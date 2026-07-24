# platform/auth

Shared auth primitives for cookies and request context.

## What lives here

- Cookie names/flags helpers
- Context accessors for current user/session id
- CSRF token read/write helpers used by middleware + identity

## Remaining tasks

- [ ] Centralize cookie option builders (Secure, HttpOnly, SameSite, Path, Domain)
- [ ] Avoid duplicating cookie logic across adapters
