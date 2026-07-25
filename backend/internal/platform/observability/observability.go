package observability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ismailtemuroglu/discord/internal/platform/config"
	"github.com/ismailtemuroglu/discord/internal/platform/logger"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ShutdownFunc flushes and stops observability providers.
type ShutdownFunc func(context.Context) error

// Providers holds active OpenTelemetry SDK providers for this process.
type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	LoggerProvider *sdklog.LoggerProvider
	cfg            *config.Config
}

// Init bootstraps tracing + OTLP logging toward SigNoz (or any OTLP collector).
// Call logger binding via Providers.Logger after Init succeeds.
func Init(ctx context.Context, cfg *config.Config) (*Providers, ShutdownFunc, error) {
	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("observability resource: %w", err)
	}

	tp, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		return nil, nil, fmt.Errorf("tracer provider: %w", err)
	}

	lp, err := newLoggerProvider(ctx, cfg, res)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, nil, fmt.Errorf("logger provider: %w", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	global.SetLoggerProvider(lp)

	providers := &Providers{
		TracerProvider: tp,
		LoggerProvider: lp,
		cfg:            cfg,
	}

	shutdown := func(shutdownCtx context.Context) error {
		var firstErr error
		if err := lp.Shutdown(shutdownCtx); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("logger provider shutdown: %w", err)
		}
		if err := tp.Shutdown(shutdownCtx); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("tracer provider shutdown: %w", err)
		}
		return firstErr
	}

	return providers, shutdown, nil
}

// Tracer returns a named tracer from the process provider.
func (p *Providers) Tracer(name string) trace.Tracer {
	if p == nil || p.TracerProvider == nil {
		return otel.Tracer(name)
	}
	return p.TracerProvider.Tracer(name)
}

// ZapCore returns a zap core that exports logs via OTLP (SigNoz).
func (p *Providers) ZapCore() zapcore.Core {
	serviceName := "discord-backend"
	if p != nil && p.cfg != nil && p.cfg.Observability.OTELServiceName != "" {
		serviceName = p.cfg.Observability.OTELServiceName
	}
	return otelzap.NewCore(
		serviceName,
		otelzap.WithLoggerProvider(p.LoggerProvider),
	)
}

// Logger builds the process zap logger bound to stdout + OTLP log export.
func (p *Providers) Logger() *zap.SugaredLogger {
	return logger.New(p.cfg, p.ZapCore())
}

func newResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.Observability.OTELServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(cfg.Env),
		),
	)
}

func newTracerProvider(ctx context.Context, cfg *config.Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Observability.OTELExporterOTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(samplerFromConfig(cfg.Observability.OTELTraceSampler)),
	)
	return tp, nil
}

func newLoggerProvider(ctx context.Context, cfg *config.Config, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(cfg.Observability.OTELExporterOTLPEndpoint),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	return lp, nil
}

func samplerFromConfig(name string) sdktrace.Sampler {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "always_off", "never":
		return sdktrace.NeverSample()
	case "always_on", "otlp", "":
		// "otlp" accepted as alias for local .env mistakes
		return sdktrace.AlwaysSample()
	default:
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
}
