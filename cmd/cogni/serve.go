package main

import (
	"os"
	"path/filepath"

	"github.com/islamborghini/cogni/internal/mcp"
	"github.com/islamborghini/cogni/internal/store"
	"github.com/spf13/cobra"
)

var serveFlags struct {
	root string
	db   string
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the MCP server over stdio",
	Long:  "Starts the Cogni MCP server, speaking the Model Context Protocol over stdio so it can be embedded in coding agents like Claude Code, Cursor, and Windsurf.",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&serveFlags.root, "root", "", "repository root (default: cwd)")
	serveCmd.Flags().StringVar(&serveFlags.db, "db", "", "override the index DB path")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	root := serveFlags.root
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = cwd
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	dbPath, err := resolveDBPath(serveFlags.db, abs)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}

	s, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer s.Close()

	srv := mcp.New(s, abs)
	return srv.ServeStdio()
}
