package main

import "github.com/spf13/cobra"

var indexFlags struct {
	stats bool
}

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Build or refresh the index for a repository",
	Long:  "Walks the given repository, parses supported source files, and writes a per-repo index to ~/.cogni/<repo-hash>/index.db. Defaults to the current directory when no path is provided.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return errNotImplemented
	},
}

func init() {
	indexCmd.Flags().BoolVar(&indexFlags.stats, "stats", false, "print indexing statistics (files, symbols, duration, db size)")
	rootCmd.AddCommand(indexCmd)
}
