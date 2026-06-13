package cmd

import (
	"todo/internal/config"
	"todo/internal/logging"

	"github.com/spf13/viper"
)

// setupLogging configures slog from Viper (file + MONAIRE_* env). Called from
// root PersistentPreRun so lifecycle logs go to stdout before any subcommand runs.
func setupLogging() {
	cfg := config.LoggingConfig{
		Level:  viper.GetString("logging.level"),
		Format: viper.GetString("logging.format"),
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}
	logging.Setup(cfg)
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	logging.Setup(cfg.Logging)
	return cfg, nil
}
