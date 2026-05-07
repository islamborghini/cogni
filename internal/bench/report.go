package bench

import (
	"fmt"
	"io"
	"time"
)

// ReportMeta carries non-derived facts that belong in the report header:
// what was measured, against what, and when. The runner populates this.
type ReportMeta struct {
	TargetRepo       string
	TargetSHA        string
	RunsPerCondition int
	GeneratedAt      time.Time
	CogniVersion     string
}

// WriteReport emits a markdown bench report to w. Pure formatting; no I/O
// other than w.Write. Callers are responsible for opening / closing files.
func WriteReport(w io.Writer, meta ReportMeta, summaries []TaskSummary) error {
	fmt.Fprintln(w, "# Cogni benchmark report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- **Target repo:** %s\n", meta.TargetRepo)
	fmt.Fprintf(w, "- **Target SHA:** `%s`\n", meta.TargetSHA)
	fmt.Fprintf(w, "- **Runs per condition:** %d\n", meta.RunsPerCondition)
	fmt.Fprintf(w, "- **Cogni version:** %s\n", meta.CogniVersion)
	fmt.Fprintf(w, "- **Generated:** %s\n", meta.GeneratedAt.UTC().Format(time.RFC3339))
	fmt.Fprintln(w)

	overall := overallReduction(summaries)
	fmt.Fprintln(w, "## Headline")
	fmt.Fprintln(w)
	if len(summaries) == 0 {
		fmt.Fprintln(w, "No tasks were scored.")
	} else {
		fmt.Fprintf(w, "Mean per-task token reduction: **%+.1f%%** across %d tasks.\n",
			overall, len(summaries))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## Per-task results")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| Task | Family | Baseline mean | Cogni mean | Reduction | Baseline pass | Cogni pass |")
	fmt.Fprintln(w, "|---|---|---:|---:|---:|---:|---:|")
	for _, s := range summaries {
		fmt.Fprintf(w, "| %s | %s | %.0f | %.0f | %+.1f%% | %d/%d | %d/%d |\n",
			s.TaskID, s.Family,
			s.Baseline.MeanTotal, s.Cogni.MeanTotal,
			s.TokenReductionPct,
			s.Baseline.PassCount, s.Baseline.Runs,
			s.Cogni.PassCount, s.Cogni.Runs,
		)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "_Methodology: see [BENCHMARK.md](./BENCHMARK.md)._")
	return nil
}

func overallReduction(s []TaskSummary) float64 {
	if len(s) == 0 {
		return 0
	}
	sum := 0.0
	for _, t := range s {
		sum += t.TokenReductionPct
	}
	return sum / float64(len(s))
}
