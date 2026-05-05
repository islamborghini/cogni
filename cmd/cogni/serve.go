package main

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the MCP server over stdio",
	Long:  "Starts the Cogni MCP server, speaking the Model Context Protocol over stdio so it can be embedded in coding agents like Claude Code, Cursor, and Windsurf.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errNotImplemented
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
