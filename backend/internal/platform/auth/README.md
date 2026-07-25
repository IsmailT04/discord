# platform/auth

Shared auth primitives for cookies and request context.

## What lives here

- Cookie names/flags helpers
- Context accessors for current user/session id
- CSRF token read/write helpers used by middleware + identity

## Remaining tasks

- [x] CSRF cookie/header names + constant-time token compare
- [x] Context accessors for current user / session id
- [x] Session cookie names (`access_token`, `session_id`) + token reader
- [ ] Centralize cookie option builders (Secure, HttpOnly, SameSite, Path, Domain)
- [ ] Avoid duplicating cookie logic across adapters
- [ ] CSRF token issue helper (set cookie) for identity login/register
