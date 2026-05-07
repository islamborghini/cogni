package bench

import "strings"

// CriterionStatus is the outcome of checking a single criterion.
type CriterionStatus string

const (
	StatusPass    CriterionStatus = "pass"
	StatusFail    CriterionStatus = "fail"
	StatusSkipped CriterionStatus = "skipped"
)

// CriterionResult is the outcome of one criterion against one run.
type CriterionResult struct {
	Criterion Criterion
	Status    CriterionStatus
	// Detail is a short human-readable note (why it skipped, what was missing).
	Detail string
}

// Score is the aggregate result for a run: per-criterion outcomes plus an
// overall pass flag. Pass means every non-skipped criterion passed AND at
// least one criterion was actually checked.
type Score struct {
	Run        RunResult
	Criteria   []CriterionResult
	Pass       bool
	NumChecked int
	NumPassed  int
}

// ScoreRun evaluates a RunResult against a task's success criteria.
// A runner-level error (run.Err != nil) yields an automatic fail with no
// criterion checks attempted.
func ScoreRun(task Task, run RunResult) Score {
	s := Score{Run: run}
	if run.Err != nil {
		s.Pass = false
		return s
	}

	allPass := true
	for _, c := range task.SuccessCriteria {
		cr := checkCriterion(c, run)
		s.Criteria = append(s.Criteria, cr)
		switch cr.Status {
		case StatusPass:
			s.NumChecked++
			s.NumPassed++
		case StatusFail:
			s.NumChecked++
			allPass = false
		case StatusSkipped:
			// not counted toward checked or passed
		}
	}
	s.Pass = allPass && s.NumChecked > 0
	return s
}

func checkCriterion(c Criterion, run RunResult) CriterionResult {
	switch {
	case c.OutputContains != "":
		if strings.Contains(run.Output, c.OutputContains) {
			return CriterionResult{Criterion: c, Status: StatusPass}
		}
		return CriterionResult{
			Criterion: c,
			Status:    StatusFail,
			Detail:    "output did not contain " + quote(c.OutputContains),
		}
	case c.FileModified != "":
		for _, f := range run.FilesModified {
			if f == c.FileModified {
				return CriterionResult{Criterion: c, Status: StatusPass}
			}
		}
		return CriterionResult{
			Criterion: c,
			Status:    StatusFail,
			Detail:    "file " + quote(c.FileModified) + " was not modified",
		}
	case c.TestsPass != "":
		return CriterionResult{
			Criterion: c,
			Status:    StatusSkipped,
			Detail:    "tests_pass not yet implemented",
		}
	case c.FunctionAdded != "":
		return CriterionResult{
			Criterion: c,
			Status:    StatusSkipped,
			Detail:    "function_added not yet implemented",
		}
	}
	return CriterionResult{
		Criterion: c,
		Status:    StatusFail,
		Detail:    "empty criterion",
	}
}

func quote(s string) string {
	return "\"" + s + "\""
}
