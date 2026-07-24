# Media context

Voice channel access and LiveKit integration (tokens only — not media relay).

## What lives here

| Path | Contents |
|------|----------|
| `domain/` | Voice membership rules, mute/deafen as *signaling* state (client + tracked state) |
| `application/` | Issue join token, track join/leave voice state, enforce Connect permission |
| `ports/` | LiveKitTokenIssuer, VoiceStateStore, PermissionChecker, EventPublisher |
| `adapters/http/` | `POST /voice/channels/{id}/token` (and voice state endpoints if needed) |
| `adapters/livekit/` | LiveKit server SDK token minting |

## Boundaries

- Media RTP/WebRTC traffic goes **client ↔ LiveKit**, never through Go
- Go owns authorization and voice roster events over Realtime WS
- LiveKit server config lives in repo `/livekit` (not here)

## Remaining tasks

- [ ] Token grant after Directory `Connect` permission check
- [ ] Room naming convention per voice channel id
- [ ] Voice state in Redis; publish `voice.state.updated`
- [ ] Screen-share allowed via LiveKit permissions in token grants
- [ ] Deafen/mute: client media + optional server-visible state for UI
- [ ] OTel: token issue latency, join/leave counters
