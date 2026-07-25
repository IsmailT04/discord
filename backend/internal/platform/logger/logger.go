package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ismailtemuroglu/discord/internal/platform/config"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey struct{}

var (
	// globalAtomicLevel allows runtime level changes via SetLevel.
	globalAtomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	// defaultLogger is used when context has no logger (never nil after init).
	defaultLogger *zap.SugaredLogger
)

func init() {
	defaultLogger = newSugared(
		zapcore.NewConsoleEncoder(devEncoderConfig()),
		zapcore.InfoLevel,
	)
}

// New builds the process logger from application config.
//
// Development / staging → colored console
// Production → JSON to stdout
//
// Pass optional extra zap cores (e.g. otelzap from observability.Providers.ZapCore)
// after OTel bootstrap so logs also export to SigNoz.
func New(cfg *config.Config, extraCores ...zapcore.Core) *zap.SugaredLogger {
	return newFromConfig(cfg, defaultLevelForEnv(cfg.Env), extraCores...)
}

// NewWithLevel is like New but forces an explicit level (debug|info|warn|error).
func NewWithLevel(cfg *config.Config, level string, extraCores ...zapcore.Core) *zap.SugaredLogger {
	return newFromConfig(cfg, level, extraCores...)
}

func newFromConfig(cfg *config.Config, level string, extraCores ...zapcore.Core) *zap.SugaredLogger {
	parsed, err := zapcore.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		parsed = zapcore.InfoLevel
	}
	globalAtomicLevel.SetLevel(parsed)

	encoder := encoderForEnv(cfg.Env)
	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), globalAtomicLevel)

	cores := make([]zapcore.Core, 0, 1+len(extraCores))
	cores = append(cores, stdoutCore)
	cores = append(cores, extraCores...)

	base := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", cfg.Observability.OTELServiceName),
			zap.String("env", cfg.Env),
			zap.String("version", cfg.ServiceVersion),
		),
	)

	defaultLogger = base.Sugar()
	return defaultLogger
}

// Default returns the process-wide sugared logger.
func Default() *zap.SugaredLogger {
	return defaultLogger
}

// Sync flushes buffered log entries. Call during graceful shutdown.
func Sync() {
	if defaultLogger == nil {
		return
	}
	_ = defaultLogger.Sync()
}

// ToContext stores the logger in ctx for request-scoped fields.
func ToContext(ctx context.Context, log *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

// FromContext returns the request logger, or the process default.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return defaultLogger
	}
	if l, ok := ctx.Value(contextKey{}).(*zap.SugaredLogger); ok && l != nil {
		return l
	}
	return defaultLogger
}

// WithContext returns a child logger enriched with zap fields.
func WithContext(ctx context.Context, fields ...zap.Field) *zap.SugaredLogger {
	log := FromContext(ctx)
	if len(fields) == 0 {
		return log
	}
	return log.Desugar().With(fields...).Sugar()
}

// WithTrace returns a logger enriched with trace_id and span_id from ctx
// when an OpenTelemetry span is active (binds logs ↔ traces in SigNoz).
func WithTrace(ctx context.Context) *zap.SugaredLogger {
	log := FromContext(ctx)
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return log
	}
	return log.With(
		"trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String(),
	)
}

// SetLevel dynamically changes the log level (debug, info, warn, error).
func SetLevel(level string) error {
	parsed, err := zapcore.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}
	from := globalAtomicLevel.Level()
	globalAtomicLevel.SetLevel(parsed)
	defaultLogger.Infow("log level changed", "from", from.String(), "to", parsed.String())
	return nil
}

func defaultLevelForEnv(env string) string {
	if env == "development" {
		return "debug"
	}
	return "info"
}

func encoderForEnv(env string) zapcore.Encoder {
	if env == "production" {
		cfg := zap.NewProductionEncoderConfig()
		cfg.TimeKey = "timestamp"
		cfg.MessageKey = "message"
		cfg.LevelKey = "level"
		cfg.CallerKey = "caller"
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
		return zapcore.NewJSONEncoder(cfg)
	}
	return zapcore.NewConsoleEncoder(devEncoderConfig())
}

func devEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg
}

func newSugared(encoder zapcore.Encoder, level zapcore.Level) *zap.SugaredLogger {
	globalAtomicLevel.SetLevel(level)
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), globalAtomicLevel)
	return zap.New(core, zap.AddCaller()).Sugar()
}
