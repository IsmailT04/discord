# Social context

Friends graph and direct / group messaging scopes.

## What lives here

| Path | Contents |
|------|----------|
| `domain/` | Friendship, DM thread, GroupDM membership rules |
| `application/` | Friend request flows, open DM, create/manage group DM |
| `ports/` | FriendshipRepository, ConversationRepository, ChatBridge (send via Chat or dedicated DM store) |
| `adapters/http/` | Friends & DM HTTP APIs |
| `adapters/persistence/` | Postgres friendships and conversation members |

## Design notes

- Prefer reusing Chat message model with `scope = channel | dm | group` **or** a thin conversation id that Chat already understands
- Friendship is bidirectional after accept; block list is a later enhancement

## Remaining tasks

- [ ] Friend request / accept / decline / remove
- [ ] List friends + pending
- [ ] 1:1 DM thread get-or-create
- [ ] Group DM create / add / remove members
- [ ] Wire message delivery through Chat + Realtime with conversation ids
- [ ] Authz: only participants can read/write
- [ ] OTel: friend-request and DM open metrics
