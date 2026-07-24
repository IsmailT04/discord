# Notification context

In-app notifications, unread-related alerts, and user preferences.

## What lives here

| Path | Contents |
|------|----------|
| `domain/` | Notification types (mention, friend request, invite, …), read/unread rules |
| `application/` | Create from domain events, list, mark read, update preferences |
| `ports/` | NotificationRepository, PreferenceRepository, EventSubscriber/Publisher |
| `adapters/http/` | Notification inbox HTTP APIs |
| `adapters/persistence/` | Postgres notifications + preferences |
| `adapters/pubsub/` | Subscribe to Chat/Social/Directory events via Redis |

## Remaining tasks

- [ ] Persist notifications for mentions, friend requests, server invites
- [ ] Inbox list + mark one/all read
- [ ] Preference flags (per type / per server later)
- [ ] Push `notification.created` on Realtime for online users
- [ ] Email/push providers — out of scope until later
- [ ] OTel: notification create/delivery metrics
