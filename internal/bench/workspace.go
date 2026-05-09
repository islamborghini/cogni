package bench

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Workspace is a fresh per-run checkout of the target repo at the pinned
// SHA. The agent runs inside Path; ModifiedFiles reports what it changed.
type Workspace struct {
	Path string
	SHA  string
}

// Prepare creates a temp dir under parentDir, clones repoURL into it, and
// checks out sha. Caller must call Cleanup when done.
func Prepare(repoURL, sha, parentDir string) (*Workspace, error) {
	if repoURL == "" {
		return nil, fmt.Errorf("workspace: repo url required")
	}
	if sha == "" || strings.HasPrefix(sha, "TBD") {
		return nil, fmt.Errorf("workspace: target sha is not pinned (got %q)", sha)
	}
	dir, err := os.MkdirTemp(parentDir, "cogni-bench-")
	if err != nil {
		return nil, fmt.Errorf("workspace: mkdir: %w", err)
	}
	if err := runGit(dir, "clone", "--quiet", repoURL, "."); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("workspace: clone: %w", err)
	}
	if err := runGit(dir, "checkout", "--quiet", sha); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("workspace: checkout %s: %w", sha, err)
	}
	return &Workspace{Path: dir, SHA: sha}, nil
}

// ModifiedFiles returns repo-relative paths of files the agent created,
// modified, or deleted relative to the pinned SHA. Untracked files are
// included; the order matches `git status --porcelain`.
func (w *Workspace) ModifiedFiles() ([]string, error) {
	out, err := captureGit(w.Path, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("workspace: status: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		// porcelain format: XY <space> path  (possibly "orig -> new" for renames)
		if len(line) < 4 {
			continue
		}
		path := line[3:]
		if i := strings.Index(path, " -> "); i != -1 {
			path = path[i+4:]
		}
		files = append(files, path)
	}
	return files, nil
}

// Diff returns the unified diff of tracked changes relative to HEAD.
func (w *Workspace) Diff() (string, error) {
	out, err := captureGit(w.Path, "diff", "HEAD", "--")
	if err != nil {
		return "", fmt.Errorf("workspace: diff: %w", err)
	}
	return out, nil
}

// Cleanup removes the workspace directory.
func (w *Workspace) Cleanup() error {
	if w.Path == "" {
		return nil
	}
	return os.RemoveAll(w.Path)
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err,
			strings.TrimSpace(stderr.String()))
	}
	return nil
}

func captureGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err,
			strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// Verify the path exists; useful in tests and for caller sanity.
func (w *Workspace) Exists() bool {
	if w.Path == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(w.Path, ".git"))
	return err == nil
}
