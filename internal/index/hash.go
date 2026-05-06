package index

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoHash returns a stable 12-char identifier for the repo at root. It
// prefers `git rev-parse --show-toplevel` + `remote.origin.url` so that
// equivalent clones at different paths share an index. Falls back to the
// absolute path when root is not a git repo.
func RepoHash(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	id := abs
	if top, err := gitOutput(abs, "rev-parse", "--show-toplevel"); err == nil {
		remote, _ := gitOutput(abs, "config", "--get", "remote.origin.url")
		id = top + "|" + remote
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])[:12], nil
}

// DBPathFor returns the per-repo SQLite path under ~/.cogni/<hash>/index.db.
func DBPathFor(root string) (string, error) {
	hash, err := RepoHash(root)
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cogni", hash, "index.db"), nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
