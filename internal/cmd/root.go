package cmd

import (
	"github.com/ekuu/dgo/internal/cmd/create"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "dgo",
	Short: "dgo tool",
}

func init() {
	rootCmd.AddCommand(create.Cmd)
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
