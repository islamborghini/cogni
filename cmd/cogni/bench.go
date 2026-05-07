package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/islamborghini/cogni/internal/bench"
	"github.com/spf13/cobra"
)

var benchFlags struct {
	tasks string
	runs  int
	list  bool
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

	// Runner integration with the Claude Code SDK / direct Anthropic API
	// lands in a follow-up PR. For now, --list is the only supported mode.
	return fmt.Errorf("bench runner not yet implemented; pass --list to inspect the task set")
}
