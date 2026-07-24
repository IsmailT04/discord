# directory/adapters

Outbound/inbound adapters implementing ports for the `directory` context.

## What lives here

One folder per technical adapter (HTTP, persistence, Redis, LiveKit, …).

## Remaining tasks

- [ ] Keep adapters thin — no business rules
- [ ] All I/O instrumented via OpenTelemetry where applicable
