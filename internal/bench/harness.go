package bench

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/islamborghini/cogni/internal/index"
	"github.com/islamborghini/cogni/internal/store"
)

// HarnessOptions configures a benchmark run.
type HarnessOptions struct {
	// RunsPerCondition is n in "n runs per (task, condition)".
	RunsPerCondition int
	// Conditions to measure. Defaults to {Baseline, Cogni}.
	Conditions []Condition
	// ParentDir for per-run workspace temp dirs. Empty = system default.
	ParentDir string
	// RunnerCfg is the Claude SDK config; reused across runs.
	RunnerCfg ClaudeRunnerConfig
	// Progress, if non-nil, receives one line per completed run.
	Progress io.Writer
}

// runOneFn executes a single (task, condition, run-index) and returns a Score.
type runOneFn func(ctx context.Context, task Task, cond Condition, runIndex int) Score

// Run drives the full benchmark: for each task × condition, prepare a fresh
// workspace, invoke the Claude SDK, score the result, and clean up. Returns
// scores for every run (failed runs included with Pass=false).
func Run(ctx context.Context, set *TaskSet, opts HarnessOptions) ([]Score, error) {
	return runLoop(ctx, set, opts, func(ctx context.Context, task Task, cond Condition, runIndex int) Score {
		return realRunOne(ctx, set, task, cond, runIndex, opts)
	})
}

func runLoop(ctx context.Context, set *TaskSet, opts HarnessOptions, run runOneFn) ([]Score, error) {
	if set == nil {
		return nil, fmt.Errorf("harness: nil task set")
	}
	if opts.RunsPerCondition <= 0 {
		opts.RunsPerCondition = 5
	}
	if len(opts.Conditions) == 0 {
		opts.Conditions = []Condition{ConditionBaseline, ConditionCogni}
	}

	var scores []Score
	for _, task := range set.Tasks {
		for _, cond := range opts.Conditions {
			for i := 0; i < opts.RunsPerCondition; i++ {
				if err := ctx.Err(); err != nil {
					return scores, err
				}
				s := run(ctx, task, cond, i)
				scores = append(scores, s)
				logProgress(opts.Progress, s)
			}
		}
	}
	return scores, nil
}

func realRunOne(ctx context.Context, set *TaskSet, task Task, cond Condition, runIndex int, opts HarnessOptions) Score {
	start := time.Now()
	ws, err := Prepare(set.TargetRepo, set.TargetSHA, opts.ParentDir)
	if err != nil {
		return Score{
			Run: RunResult{
				TaskID:     task.ID,
				Condition:  cond,
				RunIndex:   runIndex,
				Err:        fmt.Errorf("prepare workspace: %w", err),
				DurationMS: time.Since(start).Milliseconds(),
			},
		}
	}
	defer func() { _ = ws.Cleanup() }()

	if cond == ConditionCogni {
		if err := indexWorkspace(ws.Path); err != nil {
			return Score{
				Run: RunResult{
					TaskID:     task.ID,
					Condition:  cond,
					RunIndex:   runIndex,
					Err:        fmt.Errorf("index workspace: %w", err),
					DurationMS: time.Since(start).Milliseconds(),
				},
			}
		}
	}

	runner := &ClaudeRunner{Cfg: opts.RunnerCfg, Workspace: ws}
	res := runner.Run(ctx, task, cond, runIndex)
	return ScoreRun(task, res)
}

// indexWorkspace builds the cogni index for a freshly-prepared workspace so
// that the MCP server (spawned by claude via --mcp-config) sees a populated
// DB instead of returning empty results to every tool call.
func indexWorkspace(root string) error {
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
	defer s.Close()
	_, err = index.Build(root, s)
	return err
}

func logProgress(w io.Writer, s Score) {
	if w == nil {
		return
	}
	status := "PASS"
	if !s.Pass {
		status = "FAIL"
	}
	if s.Run.Err != nil {
		status = "ERR"
	}
	fmt.Fprintf(w, "[%s] %s/%s run=%d tokens=%d/%d duration=%dms\n",
		status, s.Run.TaskID, s.Run.Condition, s.Run.RunIndex,
		s.Run.InputTokens, s.Run.OutputTokens, s.Run.DurationMS)
}
