# Directory context

Servers, channels, membership, roles, permissions, and invites.

## What lives here

| Path | Contents |
|------|----------|
| `domain/` | Server, Channel, Role, Membership, Invite, permission flags, overwrite rules |
| `application/` | Create server, invite redeem, channel CRUD, role assign, permission evaluation |
| `ports/` | ServerRepository, ChannelRepository, RoleRepository, InviteRepository, PermissionService |
| `adapters/http/` | Server/channel/role/invite HTTP APIs |
| `adapters/persistence/` | Postgres mappings |

## Permission model (target)

- Role permission bitflags (view, send, manage messages/channels/roles, connect voice, …)
- Channel-level allow/deny overwrites
- Evaluation order documented and tested (member roles → overwrites → owner bypass)

## Remaining tasks

- [ ] Server create (owner membership) + list for user
- [ ] Text + voice channel types; optional categories later
- [ ] Invite codes (expiry, max uses)
- [ ] Roles, member role assignment
- [ ] Channel permission overwrites
- [ ] `Can(user, channel, action)` used by Chat/Media/Realtime
- [ ] Kick/ban, audit log hooks (later shippable update)
- [ ] OTel: permission-deny counters and span attributes (server/channel ids only)
