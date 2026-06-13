package migrate

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"todo/internal/config"
	"todo/internal/database"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func newMigrator(cfg *config.Config) (*migrate.Migrate, error) {
	if cfg.Database.Driver == "sqlite" {
		if err := database.EnsureSQLiteDir(cfg.Database.DSN); err != nil {
			return nil, fmt.Errorf("create sqlite directory: %w", err)
		}
	}

	source := fmt.Sprintf("file://%s", cfg.Migrations.Path)
	dbURL, err := migrateDatabaseURL(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return nil, err
	}
	return migrate.New(source, dbURL)
}

func migrateDatabaseURL(driver, dsn string) (string, error) {
	switch driver {
	case "mysql":
		if strings.HasPrefix(dsn, "mysql://") {
			return dsn, nil
		}
		return "mysql://" + dsn, nil
	case "sqlite":
		if strings.HasPrefix(dsn, "sqlite3://") {
			return dsn, nil
		}
		if dsn == ":memory:" || strings.Contains(dsn, ":memory:") {
			if strings.Contains(dsn, "cache=shared") {
				return "sqlite3://file?mode=memory&cache=shared", nil
			}
			return "sqlite3://file?mode=memory", nil
		}
		if strings.HasPrefix(dsn, "file:") {
			path := strings.TrimPrefix(dsn, "file:")
			if idx := strings.Index(path, "?"); idx >= 0 {
				path = path[:idx]
			}
			return "sqlite3://" + path, nil
		}
		return "sqlite3://" + dsn, nil
	default:
		return "", fmt.Errorf("unsupported migration driver %q", driver)
	}
}

func Up(cfg *config.Config) error {
	slog.Info("running migrations up", "path", cfg.Migrations.Path, "driver", cfg.Database.Driver)

	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("migrations up to date")
			return nil
		}
		slog.Error("migration up failed", "error", err)
		return err
	}
	slog.Info("migrations applied")
	return nil
}

func Down(cfg *config.Config) error {
	slog.Info("running migrations down", "path", cfg.Migrations.Path, "driver", cfg.Database.Driver)

	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to roll back")
			return nil
		}
		slog.Error("migration down failed", "error", err)
		return err
	}
	slog.Info("migration rolled back")
	return nil
}

func Version(cfg *config.Config) (uint, bool, error) {
	m, err := newMigrator(cfg)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()
	return m.Version()
}
