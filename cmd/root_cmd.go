package cmd

import (
	"fmt"
	"os"

	"github.com/sojamann/ocapi/registry"
	"github.com/spf13/cobra"
)

var flagDockerConfig string

var rootCmd = &cobra.Command{
	Use:   "ocapi",
	Short: "ocapi short desc- ....",
	Long:  "ocapi long desc- ....",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return registry.LoadCredentialsFromDockerConfig(flagDockerConfig)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&flagDockerConfig,
		"docker-config",
		"~/.docker/config.json",
		"Path to the config with the credentials",
	)
}
