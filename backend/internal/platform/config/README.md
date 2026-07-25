# platform/config

Application configuration loading.

## What lives here

- Env-based config (and optional config file later)
- Typed structs for DB, Redis, LiveKit, cookies, OTel, HTTP server

## Remaining tasks

- [x] Define required vs optional vars
- [ ] Validate durations, URLs, cookie secure flags
- [ ] Separate local / compose / k8s example env files (no secrets committed)