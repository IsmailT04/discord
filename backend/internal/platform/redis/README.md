# platform/redis

Redis client used for sessions, presence, pub/sub, rate limits.

## What lives here

- Client/pool construction
- Health ping
- Key namespace conventions (`session:`, `presence:`, `ratelimit:`, …)

## Remaining tasks

- [ ] Client + OTel instrumentation
- [ ] Document TTL conventions
- [ ] Pub/sub channel naming shared with Realtime adapters
