# deploy/k8s

Kubernetes manifests for production-like deployment practice.

## What lives here

- Deployments/Services for API (and frontend/nginx)
- ConfigMaps / Secret templates
- Ingress with WebSocket support
- Probes, resources, PDB (optional)

## Remaining tasks

- [ ] API Deployment + Service
- [ ] Ingress paths for SPA, API, WS
- [ ] External or in-cluster Postgres/Redis/LiveKit/SigNoz strategy documented
- [ ] Rollout after Update 13–14 in the project plan
- [ ] No real secrets in git — use Sealed Secrets / external secrets later
