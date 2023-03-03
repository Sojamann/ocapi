package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var FlagConfig string

var rootCmd = &cobra.Command{
	Use:   "ocapi",
	Short: "ocapi short desc- ....",
	Long:  "ocapi long desc- ....",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&FlagConfig, "config", "c", "~/.docker/config.json", "Path to the config with the credentials")
}
