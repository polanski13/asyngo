package cmd

import (
	"fmt"

	"github.com/polanski13/asyngo/gen"
	"github.com/spf13/cobra"
)

var (
	searchDirs  []string
	mainAPIFile string
	outputDir   string
	outputTypes []string
	excludes    []string
	strict      bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate AsyncAPI spec from Go source annotations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &gen.Config{
			SearchDirs:  searchDirs,
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

	initCmd.Flags().StringSliceVarP(&searchDirs, "dir", "d", []string{"."}, "directories to search")
	initCmd.Flags().StringVar(&mainAPIFile, "main", "main.go", "Go file with general API annotations")
	initCmd.Flags().StringVarP(&outputDir, "output", "o", "./docs", "output directory")
	initCmd.Flags().StringSliceVar(&outputTypes, "outputTypes", []string{"json", "yaml"}, "output types (json,yaml,go)")
	initCmd.Flags().StringSliceVar(&excludes, "exclude", nil, "exclude directories")
	initCmd.Flags().BoolVar(&strict, "strict", false, "treat warnings as errors")
}
