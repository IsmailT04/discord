# platform/middleware

HTTP middleware chain.

## What lives here

- Request ID
- Recover / panic handler
- Access logging (correlated with OTel)
- Session auth (load user into context)
- CSRF validation
- Rate limiting
- CORS (prefer same-origin via nginx — keep CORS tight)
- OTel HTTP middleware

## Remaining tasks

- [ ] Define middleware order in README once implemented
- [ ] Skip CSRF for safe methods; enforce on mutations
- [ ] Authenticated vs public route groups
