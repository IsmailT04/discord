package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ismailtemuroglu/discord/internal/identity/adapters/session"
	"github.com/ismailtemuroglu/discord/internal/platform/config"
	"github.com/ismailtemuroglu/discord/internal/platform/database"
	"github.com/ismailtemuroglu/discord/internal/platform/httpx"
	"github.com/ismailtemuroglu/discord/internal/platform/logger"
	"github.com/ismailtemuroglu/discord/internal/platform/middleware"
	"github.com/ismailtemuroglu/discord/internal/platform/observability"
	platformredis "github.com/ismailtemuroglu/discord/internal/platform/redis"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "api: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load backend/.env when present (no-op if missing).
	_ = godotenv.Load()
	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx := context.Background()

	providers, otelShutdown, err := observability.Init(ctx, cfg)
	if err != nil {
		return fmt.Errorf("observability: %w", err)
	}
	log := providers.Logger()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			log.Warnw("otel shutdown", "error", err)
		}
	}()

	db, err := database.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close()

	rdb, err := platformredis.NewCache(ctx, cfg)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Warnw("redis close", "error", err)
		}
	}()

	sessions := session.NewRedisStore(rdb.GetClient())
	limiter := middleware.NewTokenBucketLimiter(20, 40)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz(db, rdb))

	// Public stack for now. You will add RequireAuth route groups when wiring Identity.
	handler := middleware.Chain(mux,
		middleware.Recovery,
		middleware.RequestID,
		middleware.SpanTracker,
		middleware.AccessLog,
		middleware.RateLimit(limiter),
		middleware.LoadSession(sessions),
		middleware.CSRF,
	)

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Infow("api listening", "addr", addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Infow("shutdown signal", "signal", sig.String())
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	log.Infow("api stopped")
	logger.Sync()
	return nil
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleReadyz(db *database.Database, rdb *platformredis.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			_ = httpx.WriteError(w, &httpx.APIError{
				Code:    "NOT_READY",
				Message: "database unavailable",
				Status:  http.StatusServiceUnavailable,
			})
			return
		}
		if err := rdb.Ping(ctx); err != nil {
			_ = httpx.WriteError(w, &httpx.APIError{
				Code:    "NOT_READY",
				Message: "redis unavailable",
				Status:  http.StatusServiceUnavailable,
			})
			return
		}

		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
