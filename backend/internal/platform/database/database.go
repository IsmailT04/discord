package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/ismailtemuroglu/discord/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database wraps a Postgres connection pool.
type Database struct {
	pool *pgxpool.Pool
}

// New opens a pgx pool using platform config and verifies connectivity with Ping.
func New(ctx context.Context, cfg *config.Config) (*Database, error) {
	dsn := buildDSN(cfg)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.Database.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.Database.ConnMaxLifetime
	poolCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("open database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Database{pool: pool}, nil
}

// Close closes the underlying pool.
func (d *Database) Close() {
	if d == nil || d.pool == nil {
		return
	}
	d.pool.Close()
}

// Pool exposes the pgx pool for repositories/adapters.
func (d *Database) Pool() *pgxpool.Pool {
	return d.pool
}

// Ping checks database readiness.
func (d *Database) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

func buildDSN(cfg *config.Config) string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Database.User, cfg.Database.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Database.Host, cfg.Database.Port),
		Path:   cfg.Database.Name,
	}
	q := u.Query()
	q.Set("sslmode", cfg.Database.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}
