package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/sojamann/ocapi/registry"
	"github.com/spf13/cobra"
)

var flagDockerConfig string
var flagDebug bool

var rootCmd = &cobra.Command{
	Use:   "ocapi",
	Short: "ocapi short desc- ....",
	Long:  "ocapi long desc- ....",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if flagDebug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.Disabled)
		}
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
	rootCmd.PersistentFlags().BoolVar(
		&flagDebug,
		"debug",
		false,
		"Log everything...",
	)
}
