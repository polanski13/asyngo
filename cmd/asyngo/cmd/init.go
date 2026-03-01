package cmd

import (
	"fmt"

	"github.com/polanski13/asyngo/gen"
	"github.com/spf13/cobra"
)

var (
	searchDir   string
	mainAPIFile string
	outputDir   string
	outputTypes []string
	excludes    string
	strict      bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate AsyncAPI spec from Go source annotations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &gen.Config{
			SearchDir:   searchDir,
			MainAPIFile: mainAPIFile,
			OutputDir:   outputDir,
			OutputTypes: outputTypes,
			Excludes:    excludes,
			Strict:      strict,
		}

		g := gen.New()
		if err := g.Build(cfg); err != nil {
			return err
		}

		fmt.Printf("AsyncAPI spec generated in %s\n", outputDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&searchDir, "dir", "d", ".", "directories to search (comma-separated)")
	initCmd.Flags().StringVar(&mainAPIFile, "main", "main.go", "Go file with general API annotations")
	initCmd.Flags().StringVarP(&outputDir, "output", "o", "./docs", "output directory")
	initCmd.Flags().StringSliceVar(&outputTypes, "outputTypes", []string{"json", "yaml"}, "output types (json,yaml,go)")
	initCmd.Flags().StringVar(&excludes, "exclude", "", "exclude directories (comma-separated)")
	initCmd.Flags().BoolVar(&strict, "strict", false, "treat warnings as errors")
}
