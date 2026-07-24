# Chat context

Channel messages, replies, reactions, and attachments.

## What lives here

| Path | Contents |
|------|----------|
| `domain/` | Message, Reaction, Attachment metadata, reply rules |
| `application/` | Send/edit/delete message, react, attach, list with cursor pagination |
| `ports/` | MessageRepository, AttachmentStore, EventPublisher, PermissionChecker |
| `adapters/http/` | Message & attachment HTTP APIs |
| `adapters/persistence/` | Postgres messages/reactions/attachment metadata |
| `adapters/storage/` | Postgres `BYTEA` blob implementation of AttachmentStore |

## Storage policy

- No S3 for now; binary payloads in Postgres `BYTEA`
- Do **not** store base64 in the database — store raw bytes
- Enforce size and count limits; expose download via authenticated endpoint
- Keep `AttachmentStore` port so S3 can replace Postgres later

## Remaining tasks

- [ ] Message CRUD + cursor pagination
- [ ] Replies (`parent_message_id`)
- [ ] Reactions add/remove
- [ ] Multipart upload → BYTEA; download with authz
- [ ] Publish domain events to Realtime (message.created/updated/deleted, reaction.*)
- [ ] Enforce Directory permissions before every write
- [ ] OTel: message write latency, attachment size histograms
