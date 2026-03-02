package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.5.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("asyngo version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
