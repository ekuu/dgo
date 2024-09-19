package cmd

import (
	"github.com/ekuu/dgo/internal/cmd/initial"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "dgo",
	Short: "dgo tool",
}

func init() {
	rootCmd.AddCommand(initial.Cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
