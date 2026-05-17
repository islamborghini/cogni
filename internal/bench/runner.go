package bench

import "context"

// Condition identifies whether a run had Cogni's MCP tools registered.
type Condition string

const (
	ConditionBaseline Condition = "baseline"
	ConditionCogni    Condition = "cogni"
)

// RunResult captures everything the scorer needs from a single agent run.
// Runners (Claude Code SDK, direct Anthropic API, fakes) populate this.
type RunResult struct {
	TaskID    string
	Condition Condition
	RunIndex  int

	// Output is the agent's final text response — what the user would see.
	Output string

	// FilesModified lists repo-relative paths the agent edited during the
	// run. Empty for read-only tasks.
	FilesModified []string

	// Diff is the unified git diff from HEAD after the run. It is used by
	// criteria that need to inspect the shape of edits, not just file paths.
	Diff string

	// WorkspacePath is the checkout path where the run executed. It lets
	// post-run criteria execute repository-local checks such as pytest.
	WorkspacePath string

	// InputTokens / OutputTokens are summed across every API call in the
	// run, taken from API response metadata (not estimated).
	InputTokens  int
	OutputTokens int

	// CacheCreationTokens / CacheReadTokens capture Anthropic prompt-cache
	// activity. Cogni runs typically have much larger cached context (the
	// MCP tool schemas + tool results), so a token total that ignores cache
	// understates cogni's real billing footprint.
	CacheCreationTokens int
	CacheReadTokens     int

	// DurationMS is wall-clock time for the run.
	DurationMS int64

	// Err is non-nil if the runner itself failed (network, SDK crash). A
	// task-level failure (criterion mismatch) is NOT recorded here — it
	// shows up in scoring.
	Err error
}

// Runner executes one task in one condition and returns a RunResult.
// Implementations are expected to be safe for concurrent use across tasks
// but a single call is sequential.
type Runner interface {
	Run(ctx context.Context, task Task, cond Condition, runIndex int) RunResult
}
