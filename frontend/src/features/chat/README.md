# features/chat

Message list, composer, replies, reactions, attachments.

## What lives here

| Path | Contents |
|------|----------|
| `api/` | Message and attachment HTTP APIs. |
| `components/` | Message list, bubble, composer, reaction bar, reply preview. |
| `hooks/` | Infinite history, send/edit/delete, react. |
| `stores/` | Composer draft, reply target, ephemeral UI state. |

## Remaining tasks

- [ ] Virtualized message list
- [ ] Composer with reply + attachments
- [ ] Live updates via shared WS events
- [ ] Reactions UI
- [ ] Optimistic updates with reconciliation
