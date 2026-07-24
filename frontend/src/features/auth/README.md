# features/auth

Login, register, logout, session bootstrap, CSRF bootstrap.

## What lives here

| Path | Contents |
|------|----------|
| `api/` | Auth HTTP calls (login/register/refresh/logout/me). |
| `components/` | Forms and auth-only UI pieces. |
| `hooks/` | useSession, useLogin mutations. |
| `pages/` | Login and register pages. |

## Remaining tasks

- [ ] Login/register pages
- [ ] Session provider integration with shared API client
- [ ] Redirect unauthenticated users
- [ ] Logout clears query cache and WS
