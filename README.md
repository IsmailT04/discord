# Discord-like Web Application — Project Plan

Hobby project with **production-grade structure**: Discord-style servers, text/voice, social features. Not aimed at large public traffic; aimed at solid engineering practice.

---

## 1. Goals & non-goals

### Goals

- Web app (no native desktop/mobile clients)
- Text channels, voice channels, screen share, mute/deafen
- Profiles, reactions, replies, attachments
- Roles & permissions per server (incl. channel overwrites)
- Friends, DMs, group DMs, in-app notifications
- Production practices: hexagonal Go, migrations, cookies+CSRF, OTel → SigNoz, nginx LB, later Kubernetes

### Non-goals (for now)

- Real multi-tenant SaaS scale
- S3/object storage (use Postgres `BYTEA`)
- Mobile apps, Electron
- Full Discord parity (bots marketplace, Stage, Forums day one)

---

## 2. Repository layout

```
discord/
├── backend/          # Go API — DDD + hexagonal bounded contexts
├── frontend/         # React + TypeScript (Vite) SPA
├── database/         # PostgreSQL migrations + seeds
└── livekit/          # LiveKit server configs
```

Each domain/feature folder contains a `README.md` describing **what belongs there** and **remaining tasks**.

---

## 3. Technology stack

| Layer | Choice |
|-------|--------|
| Frontend | React, TypeScript, Vite |
| Backend | Go, hexagonal + DDD contexts |
| DB | PostgreSQL (`BYTEA` attachments) |
| Cache / sessions / pubsub | Redis |
| Voice / screen share | LiveKit |
| Auth | Access + refresh **HttpOnly cookies** + **CSRF** |
| Observability | OpenTelemetry → **SigNoz** (traces, metrics, logs, latency) |
| Edge | nginx reverse proxy / load balancer |
| Orchestration (later) | Kubernetes |

---

## 4. Backend bounded contexts

| Context | Owns |
|---------|------|
| **identity** | Users, sessions, CSRF, profiles |
| **directory** | Servers, channels, roles, permissions, invites |
| **chat** | Messages, replies, reactions, attachments |
| **social** | Friends, DMs, group DMs |
| **realtime** | WebSocket hub, presence, typing, fanout |
| **media** | LiveKit token issuance, voice state signaling |
| **notification** | In-app notifications & preferences |
| **platform** | Config, DB, Redis, HTTP, middleware, **OTel** |

Media RTP goes **client ↔ LiveKit**. Go never relays WebRTC media.

---

## 5. Frontend feature map

| Feature | Owns |
|---------|------|
| `auth` | Login/register/session |
| `servers` | Server rail, create/join |
| `channels` | Channel sidebar |
| `chat` | Messages, replies, reactions, attachments |
| `voice` | LiveKit room UI (mute/deafen/share) |
| `friends` | Friend graph UI |
| `dm` | DM / group DM UI |
| `notifications` | Inbox & badges |
| `profile` | Profile card/edit |
| `settings` | Preferences |
| `shared` | API client, WS, UI kit, config |
| `app` | Router, layouts, providers |

---

## 6. Cross-cutting decisions

1. **Attachments:** Postgres `BYTEA` via `AttachmentStore` port (swap to S3 later without rewriting Chat domain).
2. **Auth:** Opaque sessions in Redis (preferred); rotating refresh; CSRF on mutations; same-origin via nginx.
3. **Realtime:** App events on WS + Redis pub/sub for multi-instance; voice media on LiveKit.
4. **Permissions:** Evaluated in Directory; enforced by Chat/Media/Realtime before side effects.
5. **Observability:** OTel traces/metrics/logs from Go (and optional FE RUM later) exported OTLP to SigNoz. Capture RPS, latency histograms, error rates, WS connection gauges, auth failures.
6. **Deploy path:** Compose → nginx multi-instance → Kubernetes.

---

## 7. Shippable production plan

Each update is **demoable**. Do not start the next until the current one works end-to-end.

### Update 0 — Skeleton (current)

**Ship:** Folder architecture, inits, READMEs, plan.

- [x] Repo folders: `backend`, `frontend`, `database`, `livekit`
- [x] Go module init + hexagonal context folders + READMEs
- [x] Vite React-TS init + feature folders + READMEs
- [x] Database / LiveKit / deploy / observability README placeholders
- [ ] `npm install` in frontend (when you start coding)
- [ ] Root/tooling: Makefile, `.env.example`, `.gitignore` refinements

### Update 1 — Platform + Identity API

**Ship:** Register/login/logout/refresh with cookies + CSRF; health probes; OTel → SigNoz (local).

- Platform: config, Postgres, Redis, middleware, observability bootstrap
- Identity domain + application + adapters
- Migrations: users (+ sessions if DB-backed)
- SigNoz sees at least one HTTP trace and latency metric

### Update 2 — Frontend auth shell

**Ship:** Login/register UI, protected layout chrome, cookie API client + CSRF.

- `features/auth`, `app/router`, `app/layouts`, `shared/api`
- Session survives refresh; logout clears state

### Update 3 — Servers & channels (Directory)

**Ship:** Create server, invite join, list text channels.

- Directory migrations + APIs
- Frontend server rail + channel sidebar
- Owner vs member stub permissions

### Update 4 — Messages + WebSocket

**Ship:** Two browsers chat live in a channel.

- Chat send/list (cursor pagination)
- Realtime WS + Redis pub/sub
- Frontend chat list + composer + WS events
- OTel: message write latency, active WS connections

### Update 5 — Replies, reactions, edit/delete

**Ship:** Conversational chat UX synced live.

### Update 6 — Attachments (BYTEA)

**Ship:** Upload/download files with size limits; `AttachmentStore` port.

### Update 7 — Roles & permissions

**Ship:** Roles, assigns, channel overwrites; enforce send/view/manage; UI for basic role admin.

### Update 8 — Voice + LiveKit

**Ship:** Join voice, mute, deafen, screen share; token authz; voice roster WS events.

- Configure `/livekit`
- `internal/media` + `features/voice`

### Update 9 — Presence, typing, unread, notifications

**Ship:** Online status, typing, unread badges, in-app notification inbox.

### Update 10 — Friends, DMs, group DMs

**Ship:** Social graph + private messaging reusing chat/realtime patterns.

### Update 11 — Moderation & server ops

**Ship:** Kick/ban, pins, richer invites, audit log, channel categories/reorder; optional search.

### Update 12 — Hardening

**Ship:** Rate limits, validation, tests for authz/refresh reuse, CI, metrics/alerts in SigNoz, backup notes.

### Update 13 — nginx + multi-instance

**Ship:** SPA + `/api` + `/ws` behind nginx; ≥2 API replicas; shared Redis sessions; LiveKit separate.

### Update 14 — Kubernetes

**Ship:** Manifests under `backend/deploy/k8s`; probes; Ingress WS; secrets templates.

### Update 15 — Optional Discord parity

Threads, custom emoji, link embeds, 2FA, webhooks/bots, PTT polish — one slice at a time.

---

## 8. Critical path (first milestone)

```
Update 0 (done structurally)
  → 1 Identity + OTel/SigNoz
  → 2 Auth UI
  → 3 Servers/Channels
  → 4 Live text chat
```

**30-day target:** Updates **1–4** — login, servers, channels, realtime messaging with observability.

---

## 9. Definition of done (per update)

- Migration(s) if schema changed (`database/migrations`)
- Context/feature README task checkboxes updated
- Authz on every new resource / WS subscription
- No secrets in git
- Basic manual test: two-browser script
- Signal in SigNoz for new critical path (trace or metric) once OTel is wired (from Update 1+)

---

## 10. Observability (SigNoz) checklist

| Phase | What to see in SigNoz |
|-------|------------------------|
| Update 1 | API service traces, HTTP latency histogram, error rate |
| Update 4 | WS connection gauge, message publish latency |
| Update 6 | Attachment upload size/latency |
| Update 8 | LiveKit token issue latency |
| Update 12 | Alert rules: high error rate, auth failure spike, DB saturation |

Collector config lives in `backend/deploy/observability`. SDK wiring in `backend/internal/platform/observability`.

---

## 11. Security baseline

- HttpOnly + Secure + SameSite cookies
- CSRF on mutating cookie-auth requests
- Password hashing (argon2id recommended)
- Refresh rotation + reuse detection
- Permission checks in application services
- Attachment MIME/size limits
- Rate limit login and message send
- Never log tokens/passwords; scrub OTel attributes

---

## 12. How to use this repo (process)

1. Pick the next **Update N** from §7  
2. Open the relevant context/feature `README.md` and implement listed tasks  
3. Add SQL under `database/migrations`  
4. Touch LiveKit only for Update 8+ (`livekit/`)  
5. Keep adapters thin; keep domain free of SQL/HTTP  
6. Mark README checkboxes when done  

---

## 13. Explicitly deferred

- S3/MinIO  
- Email/push providers  
- Microservices split (single deployable API until proven need)  
- Kafka (Redis pub/sub first)  

---

*Structure and plan only — application business logic not implemented yet. Vite and Go module were initialized as scaffolding.*
