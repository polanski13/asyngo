package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

const fallbackVersion = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("asyngo version %s\n", resolveVersion())
	},
}

func resolveVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		if v != "" && v != "(devel)" {
			return v
		}
	}
	return fallbackVersion
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
