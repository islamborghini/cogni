package bench

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// resumeMetaFilename is the sidecar written into RunsDir on first use.
// It pins the invocation shape so a later rerun can either resume (all
// fields match) or refuse to silently reuse stale transcripts produced
// under different inputs.
const resumeMetaFilename = ".bench-meta.json"

type resumeMeta struct {
	TargetRepo      string `json:"target_repo"`
	TargetSHA       string `json:"target_sha"`
	Model           string `json:"model"`
	MCPConfigSHA    string `json:"mcp_config_sha,omitempty"`
	ClaudeMDSHA     string `json:"claude_md_sha,omitempty"`
	SystemPromptSHA string `json:"system_prompt_sha,omitempty"`
	TasksSHA        string `json:"tasks_sha"`
}

// computeResumeMeta builds the fingerprint for the current invocation.
// File-backed inputs (MCP config) are hashed by content so swapping the
// path but keeping content stable still resumes; changing the bytes
// invalidates.
func computeResumeMeta(set *TaskSet, opts HarnessOptions) (resumeMeta, error) {
	m := resumeMeta{
		TargetRepo: set.TargetRepo,
		TargetSHA:  set.TargetSHA,
		Model:      opts.RunnerCfg.Model,
		TasksSHA:   hashTasks(set.Tasks),
	}
	if opts.RunnerCfg.MCPConfigPath != "" {
		h, err := hashFile(opts.RunnerCfg.MCPConfigPath)
		if err != nil {
			return m, fmt.Errorf("hash mcp config: %w", err)
		}
		m.MCPConfigSHA = h
	}
	if opts.ClaudeMD != "" {
		m.ClaudeMDSHA = hashString(opts.ClaudeMD)
	}
	if opts.RunnerCfg.SystemPrompt != "" {
		m.SystemPromptSHA = hashString(opts.RunnerCfg.SystemPrompt)
	}
	return m, nil
}

// ensureResumeMeta reconciles the sidecar with the current invocation.
// If RunsDir is empty (resume disabled) returns "" with no error.
// If the sidecar is absent it writes one and returns the meta path.
// If the sidecar is present and matches it returns the path unchanged.
// If it differs, it returns an error naming the mismatched fields —
// callers should abort rather than silently mix runs of different shape.
func ensureResumeMeta(set *TaskSet, opts HarnessOptions) (string, error) {
	if opts.RunnerCfg.RunsDir == "" {
		return "", nil
	}
	if err := os.MkdirAll(opts.RunnerCfg.RunsDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(opts.RunnerCfg.RunsDir, resumeMetaFilename)
	current, err := computeResumeMeta(set, opts)
	if err != nil {
		return "", err
	}

	existing, err := readResumeMeta(path)
	if os.IsNotExist(err) {
		return path, writeResumeMeta(path, current)
	}
	if err != nil {
		return "", err
	}
	if diff := compareResumeMeta(existing, current); diff != "" {
		return "", fmt.Errorf("resume aborted: %s in %s does not match this invocation; "+
			"either rerun with a fresh --save-runs dir or delete %s to reset",
			diff, opts.RunnerCfg.RunsDir, path)
	}
	return path, nil
}

func readResumeMeta(path string) (resumeMeta, error) {
	var m resumeMeta
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal(b, &m)
}

func writeResumeMeta(path string, m resumeMeta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// compareResumeMeta returns a comma-joined list of differing field
// names, or "" if the two metas are identical.
func compareResumeMeta(a, b resumeMeta) string {
	var diffs []string
	if a.TargetRepo != b.TargetRepo {
		diffs = append(diffs, "target_repo")
	}
	if a.TargetSHA != b.TargetSHA {
		diffs = append(diffs, "target_sha")
	}
	if a.Model != b.Model {
		diffs = append(diffs, "model")
	}
	if a.MCPConfigSHA != b.MCPConfigSHA {
		diffs = append(diffs, "mcp_config")
	}
	if a.ClaudeMDSHA != b.ClaudeMDSHA {
		diffs = append(diffs, "claude_md")
	}
	if a.SystemPromptSHA != b.SystemPromptSHA {
		diffs = append(diffs, "system_prompt")
	}
	if a.TasksSHA != b.TasksSHA {
		diffs = append(diffs, "tasks")
	}
	return strings.Join(diffs, ", ")
}

func hashTasks(tasks []Task) string {
	ids := make([]string, len(tasks))
	byID := make(map[string]Task, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
		byID[t.ID] = t
	}
	sort.Strings(ids)
	h := sha256.New()
	for _, id := range ids {
		t := byID[id]
		fmt.Fprintf(h, "%s\x00%s\x00%s\x00", t.ID, t.Family, t.Prompt)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
