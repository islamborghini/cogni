package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// mcpConfigFile mirrors the shape of bench-mcp.json: {"mcpServers": {name: {command, args}}}.
type mcpConfigFile struct {
	MCPServers map[string]struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	} `json:"mcpServers"`
}

// preflightMCPBinaries reads the MCP config and prints what bench is
// about to spawn: binary path, mtime, size, self-reported version.
// Warns loudly if any binary is older than the cogni source tree it
// presumably came from — the stale-binary trap that silently bypassed
// PR #36's file_outline fix in the v2 run.
func preflightMCPBinaries(w io.Writer, mcpConfigPath string) error {
	if mcpConfigPath == "" {
		return nil
	}
	raw, err := os.ReadFile(mcpConfigPath)
	if err != nil {
		return fmt.Errorf("preflight: read mcp config: %w", err)
	}
	var cfg mcpConfigFile
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("preflight: parse mcp config: %w", err)
	}
	if len(cfg.MCPServers) == 0 {
		return nil
	}

	fmt.Fprintln(w, "MCP preflight:")
	repoSrcMtime := latestMtime(filepath.Join(findRepoRoot(), "internal/mcp"))
	for name, srv := range cfg.MCPServers {
		path := srv.Command
		st, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(w, "  [ERR] %s: cannot stat %s: %v\n", name, path, err)
			continue
		}
		ver := captureVersion(path)
		mtime := st.ModTime().UTC().Format(time.RFC3339)
		stale := !repoSrcMtime.IsZero() && st.ModTime().Before(repoSrcMtime)
		marker := "ok"
		if stale {
			marker = "STALE"
		}
		fmt.Fprintf(w, "  [%s] %s: %s\n", marker, name, path)
		fmt.Fprintf(w, "         mtime=%s size=%d version=%q\n", mtime, st.Size(), ver)
		if stale {
			fmt.Fprintf(w, "         WARNING: binary is older than internal/mcp/ source (%s).\n",
				repoSrcMtime.UTC().Format(time.RFC3339))
			fmt.Fprintf(w, "         Rebuild with: go build -o %s ./cmd/cogni\n", path)
		}
	}
	fmt.Fprintln(w)
	return nil
}

// findRepoRoot walks up from the current working directory looking for
// a go.mod file. Returns "" if not inside a Go module, which causes the
// stale-binary check to silently skip — preferable to false positives
// when bench is run from an unrelated checkout.
func findRepoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir := cwd; ; {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// latestMtime walks dir and returns the most recent file mtime found,
// or zero if dir doesn't exist (e.g. bench was run outside the repo).
func latestMtime(dir string) time.Time {
	var latest time.Time
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	})
	return latest
}

// captureVersion runs `<bin> --version` with a short timeout, returning
// the first line of stdout. Best-effort — failures fall back to "?".
func captureVersion(bin string) string {
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		return "?"
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	if line == "" {
		return "?"
	}
	return line
}
