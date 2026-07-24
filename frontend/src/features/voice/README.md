# features/voice

LiveKit room UI: join/leave, mute, deafen, screen share, participants.

## What lives here

| Path | Contents |
|------|----------|
| `api/` | Fetch LiveKit token from backend. |
| `components/` | Voice dock, participant tiles, device selectors. |
| `hooks/` | Room connect/disconnect hooks. |
| `stores/` | Mute/deafen/device selection state. |

## Remaining tasks

- [ ] Token fetch + LiveKit room connect
- [ ] Mute / deafen / screen share controls
- [ ] Voice participant list synced with WS voice.state
- [ ] Isolate re-renders from chat
