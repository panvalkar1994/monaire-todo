package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// InitViper configures env prefix and automatic env binding. Called from Cobra init and tests.
func InitViper() {
	viper.SetEnvPrefix("MONAIRE")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Migrations MigrationsConfig `mapstructure:"migrations"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

type ServerConfig struct {
	Addr    string `mapstructure:"addr"`
	GinMode string `mapstructure:"gin_mode"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type MigrationsConfig struct {
	Path string `mapstructure:"path"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Addr == "" {
		return fmt.Errorf("server.addr is required")
	}
	if c.Server.GinMode == "" {
		c.Server.GinMode = "debug"
	}
	driver := strings.ToLower(c.Database.Driver)
	if driver != "mysql" && driver != "sqlite" {
		return fmt.Errorf("database.driver must be mysql or sqlite, got %q", c.Database.Driver)
	}
	c.Database.Driver = driver
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if c.Migrations.Path == "" {
		c.Migrations.Path = "./migrations"
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
	return nil
}
