package cmd

import (
	"todo/internal/server"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		return server.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
