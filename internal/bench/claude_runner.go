// Package bench — Claude Code SDK runner. Spawns `claude -p` in headless
// stream-json mode, parses the output, and returns a RunResult. The cogni
// condition additionally registers the cogni MCP server via --mcp-config.
package bench

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
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
}

// ClaudeRunner is a Runner that drives the Claude Code SDK in headless mode
// inside a per-run Workspace. The caller is responsible for preparing and
// cleaning up the workspace.
type ClaudeRunner struct {
	Cfg       ClaudeRunnerConfig
	Workspace *Workspace
}

// Run executes the task, returning a RunResult. The workspace's
// ModifiedFiles is captured after the agent exits.
func (r *ClaudeRunner) Run(ctx context.Context, task Task, cond Condition, runIndex int) RunResult {
	res := RunResult{TaskID: task.ID, Condition: cond, RunIndex: runIndex}

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
		args = append(args, "--mcp-config", r.Cfg.MCPConfigPath)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = r.Workspace.Path
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		res.Err = fmt.Errorf("stdout pipe: %w", err)
		return res
	}
	cmd.Stderr = io.Discard

	start := time.Now()
	if err := cmd.Start(); err != nil {
		res.Err = fmt.Errorf("start claude: %w", err)
		return res
	}

	parsed, parseErr := parseStreamJSON(stdout)
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

	if mods, err := r.Workspace.ModifiedFiles(); err == nil {
		res.FilesModified = mods
	}
	return res
}

// streamEvent is the subset of fields we read from each stream-json line.
type streamEvent struct {
	Type    string         `json:"type"`
	Subtype string         `json:"subtype,omitempty"`
	Result  string         `json:"result,omitempty"`
	Usage   *streamUsage   `json:"usage,omitempty"`
	Message *streamMessage `json:"message,omitempty"`
}

type streamUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type streamMessage struct {
	Usage *streamUsage `json:"usage,omitempty"`
}

// parsedStream is what parseStreamJSON returns: the final result text plus
// total tokens taken from the result event's usage block.
type parsedStream struct {
	Result       string
	InputTokens  int
	OutputTokens int
}

// parseStreamJSON reads newline-delimited JSON events from r and returns
// the final result text plus token totals. Prefers the `result` event's
// usage; falls back to summing per-assistant usage if no result event
// arrives (e.g. the SDK was killed mid-stream).
func parseStreamJSON(r io.Reader) (parsedStream, error) {
	var out parsedStream
	var sumIn, sumOut int
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
			}
		case "assistant":
			if e.Message != nil && e.Message.Usage != nil {
				sumIn += e.Message.Usage.InputTokens
				sumOut += e.Message.Usage.OutputTokens
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return out, fmt.Errorf("read stream: %w", err)
	}
	if !sawResult {
		out.InputTokens = sumIn
		out.OutputTokens = sumOut
	}
	return out, nil
}
