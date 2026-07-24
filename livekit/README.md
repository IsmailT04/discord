# livekit

LiveKit server configuration for voice and screen sharing.

## What lives here

- `livekit.yaml` (or equivalent) server config
- API key/secret **references** (real secrets via env / k8s secrets — not committed)
- Dev run notes (Docker image, ports, Redis usage if enabled for LiveKit)
- Room defaults / webhook endpoints (optional later)

## Boundaries

- LiveKit handles WebRTC media (audio, video, screen share)
- Backend `internal/media` only mints join tokens after permission checks
- Frontend `features/voice` connects to LiveKit with that token
- Do **not** route media RTP through the Go API or nginx API upstream

## Remaining tasks

- [ ] Add `livekit.yaml` for local development
- [ ] Document required env: `LIVEKIT_URL`, `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`
- [ ] Compose service entry (referenced from `backend/deploy/compose`)
- [ ] K8s deployment notes when reaching Update 14
- [ ] Decide room naming: e.g. `voice_<channel_uuid>`
- [ ] Optional: LiveKit webhooks → backend for server-side participant truth

## Rules

- No application business logic here — config and ops docs only
- Never commit real API secrets
