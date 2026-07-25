# platform/observability

OpenTelemetry bootstrap for traces + logs → SigNoz (OTLP HTTP).

## What lives here

- `Init(ctx, cfg)` — TracerProvider + LoggerProvider, W3C propagation, global registration
- `Providers.Tracer(name)` — named tracer
- `Providers.ZapCore()` — otelzap core for stdout+OTLP tee
- `Providers.Logger()` — binds platform logger to OTel (call after `Init`)
- `ShutdownFunc` — flush providers on process exit

## Bind order in `cmd/api`

1. `config.Load()`
2. `observability.Init(ctx, cfg)` → `providers`, `shutdown`
3. `log := providers.Logger()`  // stdout + OTLP logs
4. Use `logger.WithTrace(ctx)` inside handlers/middleware when a span is active
5. On SIGTERM: `shutdown(ctx)` then `logger.Sync()`

## Remaining tasks

- [x] TracerProvider (OTLP HTTP)
- [x] LoggerProvider + otelzap bridge
- [x] Bind to platform logger
- [ ] MeterProvider / HTTP metrics
- [ ] TLS / authenticated OTLP for non-local
- [ ] Optional disable when collector is down (dev resilience)
