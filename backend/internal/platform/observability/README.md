# platform/observability

OpenTelemetry instrumentation and export to SigNoz.

## What lives here

- TracerProvider, MeterProvider, LoggerProvider bootstrap
- OTLP exporter configuration (endpoint, headers, TLS)
- Resource attributes: `service.name`, `service.version`, `deployment.environment`
- Helpers to start spans / record metrics from application code
- Propagation (W3C tracecontext) for HTTP and optionally WS

## Signals to capture

| Signal | Examples |
|--------|----------|
| Traces | HTTP handlers, use cases, DB queries, Redis, LiveKit token issue, WS event publish |
| Metrics | RPS, latency histograms, error rates, WS connections, queue/publish lag |
| Logs | Structured logs with `trace_id` / `span_id` correlation |

## Remaining tasks

- [ ] OTLP → SigNoz (see `backend/deploy/observability`)
- [ ] Auto-instrument HTTP server + pgx/redis where available
- [ ] Custom business metrics (messages sent, auth failures)
- [ ] Sampling strategy for prod vs local
- [ ] Document how to open a trace in SigNoz UI from a request id
