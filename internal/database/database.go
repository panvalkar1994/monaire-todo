package database

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"todo/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// EnsureSQLiteDir creates the parent directory for a file-based SQLite DSN.
func EnsureSQLiteDir(dsn string) error {
	if dsn == ":memory:" || strings.Contains(dsn, ":memory:") {
		return nil
	}

	path := strings.TrimPrefix(dsn, "file:")
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	if path == "" {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	case "sqlite":
		if err := EnsureSQLiteDir(cfg.DSN); err != nil {
			return nil, fmt.Errorf("create sqlite directory: %w", err)
		}
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported driver %q", cfg.Driver)
	}
	if err != nil {
		slog.Error("database connection failed", "driver", cfg.Driver, "error", err)
		return nil, err
	}

	slog.Info("database connected", "driver", cfg.Driver)
	return db, nil
}
