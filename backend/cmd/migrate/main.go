package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ismailtemuroglu/discord/internal/platform/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")

	if len(args) == 0 {
		return errors.New("usage: migrate <up|down|status|create|drop|reset|fresh|force> [args]")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	migrationsDir, err := resolveMigrationsDir()
	if err != nil {
		return err
	}

	cmd := args[0]
	switch cmd {
	case "create":
		if len(args) < 2 || args[1] == "" {
			return errors.New("usage: migrate create <name>")
		}
		return createMigration(migrationsDir, args[1])
	case "drop", "reset", "fresh", "up", "down", "status", "force":
		m, err := newMigrator(cfg, migrationsDir)
		if err != nil {
			return err
		}
		defer func() {
			srcErr, dbErr := m.Close()
			if srcErr != nil {
				fmt.Fprintf(os.Stderr, "migrate: source close: %v\n", srcErr)
			}
			if dbErr != nil {
				fmt.Fprintf(os.Stderr, "migrate: database close: %v\n", dbErr)
			}
		}()

		switch cmd {
		case "up":
			return migrateUp(m)
		case "down":
			return migrateDown(m)
		case "status":
			return migrateStatus(m)
		case "force":
			if len(args) < 2 {
				return errors.New("usage: migrate force <version>")
			}
			version, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid version %q: %w", args[1], err)
			}
			return m.Force(version)
		case "drop":
			return m.Drop()
		case "reset":
			if err := m.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			return nil
		case "fresh":
			if err := m.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			// Re-open after drop — schema_migrations is gone.
			m2, err := newMigrator(cfg, migrationsDir)
			if err != nil {
				return err
			}
			defer func() {
				_, _ = m2.Close()
			}()
			return migrateUp(m2)
		}
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
	return nil
}

func newMigrator(cfg *config.Config, migrationsDir string) (*migrate.Migrate, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	sourceURL := "file://" + filepath.ToSlash(migrationsDir)
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return nil, fmt.Errorf("open migrator: %w", err)
	}
	return m, nil
}

func migrateUp(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return migrateStatus(m)
}

func migrateDown(m *migrate.Migrate) error {
	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return migrateStatus(m)
}

func migrateStatus(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("version: none (no migrations applied)")
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Printf("version: %d  dirty: %v\n", version, dirty)
	return nil
}

func resolveMigrationsDir() (string, error) {
	candidates := []string{
		filepath.Join("..", "database", "migrations"),
		filepath.Join("database", "migrations"),
	}
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if st, err := os.Stat(abs); err == nil && st.IsDir() {
			return abs, nil
		}
	}
	return "", errors.New("could not find database/migrations (run from backend/ or repo root)")
}

func createMigration(dir, name string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	next := 1
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(e.Name(), "%d_", &n); err == nil && n >= next {
			next = n + 1
		}
	}

	up := filepath.Join(dir, fmt.Sprintf("%06d_%s.up.sql", next, name))
	down := filepath.Join(dir, fmt.Sprintf("%06d_%s.down.sql", next, name))
	for _, path := range []string{up, down} {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_ = f.Close()
		fmt.Println("created", path)
	}
	return nil
}
