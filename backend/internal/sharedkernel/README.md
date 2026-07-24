# sharedkernel

Tiny cross-context primitives only.

## What lives here

- ID types / generators conventions
- Pagination cursor types
- Shared time/optional helpers that are not platform infra

## Rules

- **Do not** put business rules here
- Prefer duplication over wrong abstractions until a real shared need appears

## Remaining tasks

- [ ] Agree ID strategy (ULID/UUID) and stick to it in migrations
- [ ] Cursor pagination opaque encoding helpers if shared by Chat/Social
