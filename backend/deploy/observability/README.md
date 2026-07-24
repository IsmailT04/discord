# deploy/observability

SigNoz + OpenTelemetry Collector configuration.

## What lives here

- OTel Collector config (receivers: OTLP; exporters: SigNoz)
- Compose/Helm notes or values for SigNoz
- Dashboard / alert ideas (latency, error rate, WS connections)

## Pipeline (target)

```
Go API / (future FE RUM)  --OTLP-->  OTel Collector  -->  SigNoz
```

## Remaining tasks

- [ ] Collector config for traces, metrics, logs
- [ ] SigNoz local run instructions (compose or upstream chart)
- [ ] Align service names with `platform/observability`
- [ ] Example queries: p95 HTTP latency, auth error rate, WS gauge
- [ ] Retention and sampling notes for hobby prod
