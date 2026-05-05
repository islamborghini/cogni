package main

import "github.com/spf13/cobra"

var benchFlags struct {
	tasks string
	runs  int
}

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run the token-savings benchmark",
	Long:  "Runs a fixed task set against a fixed commit of a fixed repository, with and without Cogni's tools available, and reports tokens used and pass/fail per task.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errNotImplemented
	},
}

func init() {
	benchCmd.Flags().StringVar(&benchFlags.tasks, "tasks", "internal/bench/tasks.yaml", "path to the benchmark task definition file")
	benchCmd.Flags().IntVar(&benchFlags.runs, "runs", 5, "number of runs per task per condition")
	rootCmd.AddCommand(benchCmd)
}
