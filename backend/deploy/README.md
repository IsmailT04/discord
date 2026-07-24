# deploy

Deployment and local infrastructure manifests for the backend stack.

## Layout

| Path | Purpose |
|------|---------|
| `compose/` | Docker Compose for API + deps (Postgres, Redis, LiveKit ref, SigNoz/OTel) |
| `nginx/` | Reverse proxy / load balancer configs |
| `k8s/` | Kubernetes manifests (Deployments, Services, Ingress, Secrets templates) |
| `observability/` | OpenTelemetry Collector + SigNoz connection config |

## Remaining tasks

- [ ] Compose stack for local prod-like run
- [ ] nginx: SPA + `/api` + `/ws` routing, WebSocket timeouts
- [ ] K8s after multi-instance works on compose
- [ ] Document secret management (do not commit real secrets)
