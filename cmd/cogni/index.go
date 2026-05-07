package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/islamborghini/cogni/internal/index"
	"github.com/islamborghini/cogni/internal/store"
	"github.com/spf13/cobra"
)

var indexFlags struct {
	stats bool
	db    string
}

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Build or refresh the index for a repository",
	Long:  "Walks the given repository, parses supported source files, and writes a per-repo index to ~/.cogni/<repo-hash>/index.db. Defaults to the current directory when no path is provided.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runIndex,
}

func init() {
	indexCmd.Flags().BoolVar(&indexFlags.stats, "stats", false, "print indexing statistics (files, symbols, duration, db size)")
	indexCmd.Flags().StringVar(&indexFlags.db, "db", "", "override the index DB path (default ~/.cogni/index.db)")
	rootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) == 1 {
		root = args[0]
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	dbPath, err := resolveDBPath(indexFlags.db, abs)
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

	stats, err := index.Build(abs, s)
	if err != nil {
		return err
	}

	if indexFlags.stats {
		printStats(cmd, abs, dbPath, stats)
	}
	return nil
}

func resolveDBPath(override, root string) (string, error) {
	if override != "" {
		return override, nil
	}
	return index.DBPathFor(root)
}

func printStats(cmd *cobra.Command, root, dbPath string, st index.Stats) {
	w := cmd.OutOrStdout()
	size := dbSize(dbPath)
	fmt.Fprintf(w, "root:      %s\n", root)
	fmt.Fprintf(w, "db:        %s (%s)\n", dbPath, humanBytes(size))
	fmt.Fprintf(w, "scanned:   %d\n", st.FilesScanned)
	fmt.Fprintf(w, "indexed:   %d\n", st.FilesIndexed)
	fmt.Fprintf(w, "skipped:   %d\n", st.FilesSkipped)
	fmt.Fprintf(w, "symbols:   %d\n", st.Symbols)
	fmt.Fprintf(w, "refs:      %d\n", st.Refs)
	fmt.Fprintf(w, "duration:  %s\n", st.Duration.Round(1e6))
	if n := len(st.Errors); n > 0 {
		fmt.Fprintf(w, "errors:    %d (showing up to 5)\n", n)
		for i, e := range st.Errors {
			if i >= 5 {
				break
			}
			fmt.Fprintf(w, "  - %v\n", e)
		}
	}
}

func dbSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
