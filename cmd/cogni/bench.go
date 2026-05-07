package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/islamborghini/cogni/internal/bench"
	"github.com/spf13/cobra"
)

var benchFlags struct {
	tasks     string
	runs      int
	list      bool
	report    string
	mcpConfig string
	model     string
	claudeCmd string
	saveRuns  string
}

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run the token-savings benchmark",
	Long: "Runs a fixed task set against a fixed commit of a fixed repository, " +
		"with and without Cogni's tools available, and reports tokens used and " +
		"pass/fail per task. See BENCHMARK.md for methodology.",
	RunE: runBench,
}

func init() {
	benchCmd.Flags().StringVar(&benchFlags.tasks, "tasks", "internal/bench/tasks.yaml", "path to the benchmark task definition file")
	benchCmd.Flags().IntVar(&benchFlags.runs, "runs", 5, "number of runs per task per condition")
	benchCmd.Flags().BoolVar(&benchFlags.list, "list", false, "load and list tasks without running them")
	benchCmd.Flags().StringVar(&benchFlags.report, "report", "bench-report.md", "path to write the markdown report")
	benchCmd.Flags().StringVar(&benchFlags.mcpConfig, "mcp-config", "", "path to MCP config registering cogni (used in the cogni condition)")
	benchCmd.Flags().StringVar(&benchFlags.model, "model", "", "override Claude model (default: SDK default)")
	benchCmd.Flags().StringVar(&benchFlags.claudeCmd, "claude-cmd", "claude", "path to the claude CLI")
	benchCmd.Flags().StringVar(&benchFlags.saveRuns, "save-runs", "", "directory to write per-run stream-json transcripts")
	rootCmd.AddCommand(benchCmd)
}

func runBench(cmd *cobra.Command, args []string) error {
	set, err := bench.LoadTasks(benchFlags.tasks)
	if err != nil {
		return err
	}

	if benchFlags.list {
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tFAMILY\tCRITERIA\n")
		for _, t := range set.Tasks {
			fmt.Fprintf(w, "%s\t%s\t%d\n", t.ID, t.Family, len(t.SuccessCriteria))
		}
		return w.Flush()
	}

	opts := bench.HarnessOptions{
		RunsPerCondition: benchFlags.runs,
		Progress:         cmd.ErrOrStderr(),
		RunnerCfg: bench.ClaudeRunnerConfig{
			ClaudeCmd:     benchFlags.claudeCmd,
			MCPConfigPath: benchFlags.mcpConfig,
			Model:         benchFlags.model,
			RunsDir:       benchFlags.saveRuns,
		},
	}

	scores, err := bench.Run(cmd.Context(), set, opts)
	if err != nil {
		return fmt.Errorf("bench run: %w", err)
	}

	summaries := bench.Aggregate(set, scores)

	f, err := os.Create(benchFlags.report)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	defer f.Close()
	meta := bench.ReportMeta{
		TargetRepo:       set.TargetRepo,
		TargetSHA:        set.TargetSHA,
		RunsPerCondition: benchFlags.runs,
		GeneratedAt:      time.Now(),
		CogniVersion:     version,
	}
	if err := bench.WriteReport(f, meta, summaries); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Report written to %s (%d tasks)\n", benchFlags.report, len(summaries))
	return nil
}
