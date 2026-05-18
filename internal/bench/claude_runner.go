// Package bench — Claude Code SDK runner. Spawns `claude -p` in headless
// stream-json mode, parses the output, and returns a RunResult. The cogni
// condition additionally registers the cogni MCP server via --mcp-config.
package bench

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeRunnerConfig configures the Claude Code subprocess runner.
type ClaudeRunnerConfig struct {
	// ClaudeCmd is the path to the `claude` CLI. Defaults to "claude" (PATH lookup).
	ClaudeCmd string
	// MCPConfigPath, when non-empty, is passed via --mcp-config in the
	// cogni condition. Baseline runs ignore it.
	MCPConfigPath string
	// Model overrides the model selection. Empty means SDK default.
	Model string
	// RunsDir, if non-empty, receives a per-run JSONL transcript named
	// "<task>-<condition>-<run>.jsonl" — the raw stream-json output.
	RunsDir string
	// SystemPrompt, if non-empty, is appended via --append-system-prompt
	// in the cogni condition only. Lets the harness simulate a CLAUDE.md
	// nudge that steers the agent toward cogni's tools.
	SystemPrompt string
}

// ClaudeRunner is a Runner that drives the Claude Code SDK in headless mode
// inside a per-run Workspace. The caller is responsible for preparing and
// cleaning up the workspace.
type ClaudeRunner struct {
	Cfg       ClaudeRunnerConfig
	Workspace *Workspace
}

// Run executes the task, returning a RunResult. The workspace's modified file
// list and diff are captured after the agent exits.
func (r *ClaudeRunner) Run(ctx context.Context, task Task, cond Condition, runIndex int) RunResult {
	res := RunResult{
		TaskID:        task.ID,
		Condition:     cond,
		RunIndex:      runIndex,
		WorkspacePath: r.Workspace.Path,
	}

	cmdName := r.Cfg.ClaudeCmd
	if cmdName == "" {
		cmdName = "claude"
	}

	args := []string{
		"-p", task.Prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--permission-mode", "bypassPermissions",
	}
	if r.Cfg.Model != "" {
		args = append(args, "--model", r.Cfg.Model)
	}
	if cond == ConditionCogni && r.Cfg.MCPConfigPath != "" {
		mcpPath := r.Cfg.MCPConfigPath
		if !filepath.IsAbs(mcpPath) {
			if abs, err := filepath.Abs(mcpPath); err == nil {
				mcpPath = abs
			}
		}
		args = append(args, "--mcp-config", mcpPath)
	}
	if cond == ConditionCogni && r.Cfg.SystemPrompt != "" {
		args = append(args, "--append-system-prompt", r.Cfg.SystemPrompt)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = r.Workspace.Path
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		res.Err = fmt.Errorf("stdout pipe: %w", err)
		return res
	}
	cmd.Stderr = io.Discard

	var streamReader io.Reader = stdout
	if r.Cfg.RunsDir != "" {
		if err := os.MkdirAll(r.Cfg.RunsDir, 0o755); err != nil {
			res.Err = fmt.Errorf("create runs dir: %w", err)
			return res
		}
		fname := fmt.Sprintf("%s-%s-%d.jsonl", task.ID, cond, runIndex)
		f, err := os.Create(filepath.Join(r.Cfg.RunsDir, fname))
		if err != nil {
			res.Err = fmt.Errorf("create transcript: %w", err)
			return res
		}
		defer f.Close()
		streamReader = io.TeeReader(stdout, f)
	}

	start := time.Now()
	if err := cmd.Start(); err != nil {
		res.Err = fmt.Errorf("start claude: %w", err)
		return res
	}

	parsed, parseErr := parseStreamJSON(streamReader)
	waitErr := cmd.Wait()
	res.DurationMS = time.Since(start).Milliseconds()

	if parseErr != nil {
		res.Err = parseErr
		return res
	}
	if waitErr != nil && parsed.Result == "" {
		res.Err = fmt.Errorf("claude exited: %w", waitErr)
		return res
	}

	res.Output = parsed.Result
	res.InputTokens = parsed.InputTokens
	res.OutputTokens = parsed.OutputTokens
	res.CacheCreationTokens = parsed.CacheCreationInputTokens
	res.CacheReadTokens = parsed.CacheReadInputTokens

	if strings.Contains(parsed.Result, rateLimitMarker) {
		res.Err = ErrRateLimited
		return res
	}

	if mods, err := r.Workspace.ModifiedFiles(); err == nil {
		res.FilesModified = mods
	}
	if diff, err := r.Workspace.Diff(); err == nil {
		res.Diff = diff
	}
	return res
}

// ErrRateLimited indicates the Claude CLI returned the subscription
// rate-limit banner as its final result. The run may have consumed real
// tokens before being cut off, but produced no usable answer, so it
// should be treated as a transient runner error (and re-run after the
// limit resets).
var ErrRateLimited = errors.New("claude rate limit hit")

// rateLimitMarker is the substring the Claude CLI emits in its final
// result event when the subscription quota is exhausted, e.g.
// "You've hit your limit · resets 7:10pm (Asia/Almaty)".
const rateLimitMarker = "You've hit your limit"

// streamEvent is the subset of fields we read from each stream-json line.
type streamEvent struct {
	Type    string         `json:"type"`
	Subtype string         `json:"subtype,omitempty"`
	Result  string         `json:"result,omitempty"`
	Usage   *streamUsage   `json:"usage,omitempty"`
	Message *streamMessage `json:"message,omitempty"`
}

type streamUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type streamMessage struct {
	Usage *streamUsage `json:"usage,omitempty"`
}

// parsedStream is what parseStreamJSON returns: the final result text plus
// total tokens taken from the result event's usage block.
type parsedStream struct {
	Result                   string
	InputTokens              int
	OutputTokens             int
	CacheCreationInputTokens int
	CacheReadInputTokens     int
}

// ReplayTranscript parses a saved stream-json transcript and returns a
// RunResult equivalent to what Run would have produced. It is the
// read-side counterpart used by the harness's resume path: if a cell's
// transcript already exists and didn't hit the rate limit, the harness
// reuses it instead of paying to re-run.
//
// Diff and FilesModified are not recoverable from the transcript, so
// criteria depending on those will be marked failed/skipped on replay.
// For purely output-driven tasks (e.g. output_contains) replay is
// faithful.
func ReplayTranscript(path string, task Task, cond Condition, runIndex int) (RunResult, error) {
	res := RunResult{TaskID: task.ID, Condition: cond, RunIndex: runIndex}
	f, err := os.Open(path)
	if err != nil {
		return res, err
	}
	defer f.Close()
	parsed, err := parseStreamJSON(f)
	if err != nil {
		return res, err
	}
	res.Output = parsed.Result
	res.InputTokens = parsed.InputTokens
	res.OutputTokens = parsed.OutputTokens
	res.CacheCreationTokens = parsed.CacheCreationInputTokens
	res.CacheReadTokens = parsed.CacheReadInputTokens
	if strings.Contains(parsed.Result, rateLimitMarker) {
		res.Err = ErrRateLimited
	}
	return res, nil
}

// parseStreamJSON reads newline-delimited JSON events from r and returns
// the final result text plus token totals. Prefers the `result` event's
// usage; falls back to summing per-assistant usage if no result event
// arrives (e.g. the SDK was killed mid-stream).
func parseStreamJSON(r io.Reader) (parsedStream, error) {
	var out parsedStream
	var sumIn, sumOut, sumCacheCreate, sumCacheRead int
	var sawResult bool

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e streamEvent
		if err := json.Unmarshal(line, &e); err != nil {
			// One malformed line shouldn't kill the whole run.
			continue
		}
		switch e.Type {
		case "result":
			sawResult = true
			out.Result = e.Result
			if e.Usage != nil {
				out.InputTokens = e.Usage.InputTokens
				out.OutputTokens = e.Usage.OutputTokens
				out.CacheCreationInputTokens = e.Usage.CacheCreationInputTokens
				out.CacheReadInputTokens = e.Usage.CacheReadInputTokens
			}
		case "assistant":
			if e.Message != nil && e.Message.Usage != nil {
				sumIn += e.Message.Usage.InputTokens
				sumOut += e.Message.Usage.OutputTokens
				sumCacheCreate += e.Message.Usage.CacheCreationInputTokens
				sumCacheRead += e.Message.Usage.CacheReadInputTokens
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return out, fmt.Errorf("read stream: %w", err)
	}
	if !sawResult {
		out.InputTokens = sumIn
		out.OutputTokens = sumOut
		out.CacheCreationInputTokens = sumCacheCreate
		out.CacheReadInputTokens = sumCacheRead
	}
	return out, nil
}
