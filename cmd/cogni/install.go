package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/islamborghini/cogni/internal/index"
	"github.com/islamborghini/cogni/internal/store"
	"github.com/spf13/cobra"
)

const claudeMDMarkerStart = "<!-- cogni-start -->"
const claudeMDMarkerEnd = "<!-- cogni-end -->"

var installFlags struct {
	root  string
	local bool
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Configure Claude Code to use Cogni in the current repo",
	Long: `Registers Cogni in your Claude Code MCP config, writes a CLAUDE.md
with tool-selection instructions, and indexes the repository.

Run once per repository. Safe to re-run (idempotent).`,
	RunE: runInstall,
}

func init() {
	installCmd.Flags().StringVar(&installFlags.root, "root", ".", "path to the repository to configure")
	installCmd.Flags().BoolVar(&installFlags.local, "local", false, "write MCP config to .mcp.json instead of ~/.claude.json")
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	root, err := filepath.Abs(installFlags.root)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()

	var mcpConfigPath string
	if installFlags.local {
		mcpConfigPath = ".mcp.json"
	} else {
		mcpConfigPath, err = globalClaudeJSON()
		if err != nil {
			return err
		}
	}
	if err := mergeMCPConfig(mcpConfigPath, root); err != nil {
		return fmt.Errorf("write MCP config: %w", err)
	}
	fmt.Fprintf(out, "ok  registered in %s\n", mcpConfigPath)

	claudeMDPath := filepath.Join(root, "CLAUDE.md")
	if err := upsertClaudeMD(claudeMDPath); err != nil {
		return fmt.Errorf("write CLAUDE.md: %w", err)
	}
	fmt.Fprintf(out, "ok  wrote %s\n", claudeMDPath)

	fmt.Fprintln(out, "    indexing repository...")
	dbPath, err := index.DBPathFor(root)
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
	defer func() { _ = s.Close() }()
	stats, err := index.Build(root, s)
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}
	fmt.Fprintf(out, "ok  indexed %d symbols across %d files\n", stats.Symbols, stats.FilesIndexed)

	fmt.Fprintln(out, "\nRestart Claude Code to activate Cogni.")
	return nil
}

func globalClaudeJSON() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude.json"), nil
}

// mergeMCPConfig adds the cogni server entry to the JSON config file without
// disturbing other keys. Creates the file if it does not exist.
func mergeMCPConfig(path, root string) error {
	raw := []byte("{}")
	if data, err := os.ReadFile(path); err == nil {
		raw = data
	}

	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("parse existing config: %w", err)
	}

	servers := map[string]json.RawMessage{}
	if existing, ok := cfg["mcpServers"]; ok {
		if err := json.Unmarshal(existing, &servers); err != nil {
			return fmt.Errorf("parse mcpServers: %w", err)
		}
	}

	entry, err := json.Marshal(map[string]any{
		"command": "cogni",
		"args":    []string{"serve", "--root", root},
	})
	if err != nil {
		return err
	}
	servers["cogni"] = entry

	serversJSON, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	cfg["mcpServers"] = serversJSON

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// upsertClaudeMD writes or replaces the cogni block in CLAUDE.md using
// HTML comment markers so re-runs are idempotent and manual edits outside
// the markers are preserved.
func upsertClaudeMD(path string) error {
	block := claudeMDMarkerStart + "\n" + defaultClaudeMD + claudeMDMarkerEnd + "\n"

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(data)
	start := strings.Index(content, claudeMDMarkerStart)
	end := strings.Index(content, claudeMDMarkerEnd)

	switch {
	case start != -1 && end != -1:
		content = content[:start] + block + content[end+len(claudeMDMarkerEnd)+1:]
	case len(content) == 0:
		content = block
	default:
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + block
	}

	return os.WriteFile(path, []byte(content), 0o644)
}
