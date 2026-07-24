# Identity context

Authentication, sessions, CSRF, and user profiles.

## What lives here

| Path | Contents |
|------|----------|
| `domain/` | User, credentials rules, session lifecycle invariants |
| `application/` | Register, login, logout, refresh, get/update profile, CSRF issue/validate |
| `ports/` | UserRepository, SessionStore, PasswordHasher, Clock |
| `adapters/http/` | Auth & profile HTTP handlers |
| `adapters/persistence/` | Postgres user storage |
| `adapters/session/` | Redis (preferred) or Postgres session + refresh-token store |

## Cookie / CSRF design (target)

- HttpOnly, Secure, SameSite access + refresh cookies
- Short-lived access session; rotating refresh with reuse detection
- CSRF token required on mutating requests for cookie-authenticated SPA

## Remaining tasks

- [ ] User aggregate + password hashing (argon2id recommended)
- [ ] Session create/rotate/revoke; logout-all
- [ ] CSRF synchronizer or double-submit implementation
- [ ] `/auth/register`, `/auth/login`, `/auth/logout`, `/auth/refresh`, `/users/me`
- [ ] OTel spans around auth flows (no password/PII in attributes)
- [ ] Rate-limit login/register via Redis
- [ ] Unit tests: refresh reuse detection, CSRF rejection
