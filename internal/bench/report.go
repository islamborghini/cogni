package bench

import (
	"fmt"
	"io"
	"strings"
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
func WriteReport(w io.Writer, meta ReportMeta, summaries []TaskSummary, scores ...Score) error {
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
	writeFailures(w, scores)
	fmt.Fprintln(w, "_Methodology: see [BENCHMARK.md](./BENCHMARK.md)._")
	return nil
}

func writeFailures(w io.Writer, scores []Score) {
	var failures []Score
	for _, s := range scores {
		if !s.Pass {
			failures = append(failures, s)
		}
	}
	if len(failures) == 0 {
		return
	}

	fmt.Fprintln(w, "## Failed runs")
	fmt.Fprintln(w)
	for _, s := range failures {
		fmt.Fprintf(w, "- `%s/%s/%d`", s.Run.TaskID, s.Run.Condition, s.Run.RunIndex)
		if s.Run.Err != nil {
			fmt.Fprintf(w, ": runner error: %s\n", s.Run.Err)
			continue
		}
		details := criterionFailureDetails(s.Criteria)
		if len(details) == 0 {
			fmt.Fprintln(w, ": no criteria passed")
			continue
		}
		fmt.Fprintf(w, ": %s\n", strings.Join(details, "; "))
	}
	fmt.Fprintln(w)
}

func criterionFailureDetails(criteria []CriterionResult) []string {
	var out []string
	for _, cr := range criteria {
		switch cr.Status {
		case StatusFail:
			out = append(out, "FAIL "+criterionName(cr.Criterion)+detailSuffix(cr.Detail))
		case StatusSkipped:
			out = append(out, "SKIP "+criterionName(cr.Criterion)+detailSuffix(cr.Detail))
		}
	}
	return out
}

func criterionName(c Criterion) string {
	switch {
	case c.OutputContains != "":
		return "output_contains=" + quote(c.OutputContains)
	case c.FileModified != "":
		return "file_modified=" + quote(c.FileModified)
	case c.TestsPass != "":
		return "tests_pass=" + quote(c.TestsPass)
	case c.FunctionAdded != "":
		return "function_added=" + quote(c.FunctionAdded)
	default:
		return "empty"
	}
}

func detailSuffix(detail string) string {
	if detail == "" {
		return ""
	}
	return " (" + detail + ")"
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
