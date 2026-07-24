# Realtime context

WebSocket gateway, presence, typing, and cross-instance event fanout.

## What lives here

| Path | Contents |
|------|----------|
| `domain/` | Connection, Presence status, subscription (server/channel/dm), event envelope |
| `application/` | Authenticate WS, subscribe/unsubscribe, broadcast, presence heartbeat, typing |
| `ports/` | ConnectionHub, PubSub, SessionValidator, PermissionChecker |
| `adapters/websocket/` | Gorilla/nhooyr (or similar) WS server adapter |
| `adapters/pubsub/` | Redis pub/sub (or Redis Streams) for multi-instance fanout |

## Event envelope (target)

Stable event names consumed by the SPA, e.g.:

- `message.created|updated|deleted`
- `reaction.added|removed`
- `presence.updated`
- `typing.started`
- `voice.state.updated`
- `notification.created`

## Remaining tasks

- [ ] WS upgrade with cookie/session auth
- [ ] Channel/server/DM subscription with permission checks
- [ ] Local hub + Redis pub/sub bridge
- [ ] Presence online/idle/offline via heartbeat + TTL in Redis
- [ ] Typing indicators (ephemeral, not persisted)
- [ ] Reconnect guidance (resume/last_event_id optional later)
- [ ] OTel: active connections gauge, event publish latency, disconnect reasons
