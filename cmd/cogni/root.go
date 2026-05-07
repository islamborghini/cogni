package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is overwritten at build time via -ldflags "-X main.version=...".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "cogni",
	Short:   "MCP server that cuts coding-agent token usage",
	Long:    "Cogni is an open-core MCP server that drops into AI coding agents and reduces token consumption by giving them structured recall over a repository.",
	Version: version,
	// Cobra prints usage on any RunE error by default. For a real CLI that
	// is noisy — only print usage on argument-parsing errors.
	SilenceUsage: true,
}

// Execute runs the root command. It writes errors to stderr and exits with a
// non-zero status on failure so callers (shells, CI) can detect problems.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "cogni:", err)
		os.Exit(1)
	}
}
