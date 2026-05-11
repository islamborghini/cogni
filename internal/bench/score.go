package bench

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

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
		return checkTestsPass(c, run)
	case c.FunctionAdded != "":
		return checkFunctionAdded(c, run)
	}
	return CriterionResult{
		Criterion: c,
		Status:    StatusFail,
		Detail:    "empty criterion",
	}
}

var runTestsPass = runPytestCriterion
var runCommand = runBenchCommand

func checkTestsPass(c Criterion, run RunResult) CriterionResult {
	if strings.TrimSpace(c.TestsPass) == "" {
		return CriterionResult{
			Criterion: c,
			Status:    StatusFail,
			Detail:    "tests_pass target is empty",
		}
	}
	if run.WorkspacePath == "" {
		return CriterionResult{
			Criterion: c,
			Status:    StatusFail,
			Detail:    "workspace path unavailable for tests_pass",
		}
	}
	out, err := runTestsPass(run.WorkspacePath, c.TestsPass)
	if err != nil {
		return CriterionResult{
			Criterion: c,
			Status:    StatusFail,
			Detail:    "pytest failed: " + compactOutput(out, err),
		}
	}
	return CriterionResult{Criterion: c, Status: StatusPass}
}

func checkFunctionAdded(c Criterion, run RunResult) CriterionResult {
	name := functionName(c.FunctionAdded)
	if name == "" {
		return CriterionResult{
			Criterion: c,
			Status:    StatusFail,
			Detail:    "function_added target is empty",
		}
	}
	if diffAddsFunction(run.Diff, name) {
		return CriterionResult{Criterion: c, Status: StatusPass}
	}
	return CriterionResult{
		Criterion: c,
		Status:    StatusFail,
		Detail:    "diff did not add function " + quote(name),
	}
}

func functionName(qualified string) string {
	qualified = strings.TrimSpace(qualified)
	if qualified == "" {
		return ""
	}
	parts := strings.Split(qualified, ".")
	return parts[len(parts)-1]
}

func diffAddsFunction(diff, name string) bool {
	pattern := `^\+\s*(?:async\s+)?def\s+` + regexp.QuoteMeta(name) + `\s*\(`
	re := regexp.MustCompile(pattern)
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+++") {
			continue
		}
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

func runPytestCriterion(workspace, node string) (string, error) {
	python, err := ensureTestVenv(workspace)
	if err != nil {
		return "", err
	}
	return runBenchCommand(10*time.Minute, workspace, python, "-m", "pytest", node, "-x", "--tb=no", "-q")
}

func ensureTestVenv(workspace string) (string, error) {
	venv := filepath.Join(workspace, ".cogni-test-venv")
	python := venvPython(venv)
	marker := filepath.Join(venv, ".cogni-deps-installed")
	if _, err := os.Stat(marker); err == nil {
		return python, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if _, err := os.Stat(python); os.IsNotExist(err) {
		if out, err := runCommand(5*time.Minute, workspace, "python3", "-m", "venv", ".cogni-test-venv"); err != nil {
			return "", fmt.Errorf("create test venv: %w: %s", err, compactString(out))
		}
	} else if err != nil {
		return "", err
	}
	// Some repos (e.g. httpx) declare test deps in requirements.txt rather
	// than a [test] extra. Try requirements.txt first, fall back to .[test].
	reqFile := filepath.Join(workspace, "requirements.txt")
	var installOut string
	var installErr error
	if _, err := os.Stat(reqFile); err == nil {
		installOut, installErr = runCommand(10*time.Minute, workspace, python, "-m", "pip", "install", "-q", "-r", "requirements.txt")
	} else {
		installOut, installErr = runCommand(10*time.Minute, workspace, python, "-m", "pip", "install", "-q", "-e", ".[test]")
	}
	if installErr != nil {
		return "", fmt.Errorf("install test dependencies: %w: %s", installErr, compactString(installOut))
	}
	if err := os.WriteFile(marker, []byte("ok\n"), 0o644); err != nil {
		return "", fmt.Errorf("write test venv marker: %w", err)
	}
	return python, nil
}

func venvPython(venv string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venv, "Scripts", "python.exe")
	}
	return filepath.Join(venv, "bin", "python")
}

func runBenchCommand(timeout time.Duration, dir, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return out.String(), ctx.Err()
	}
	return out.String(), err
}

func compactOutput(out string, err error) string {
	msg := compactString(out)
	if msg == "" {
		return err.Error()
	}
	return err.Error() + ": " + msg
}

func compactString(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	msg := strings.Join(fields, " ")
	if len(msg) > 300 {
		return msg[:300] + "..."
	}
	return msg
}

func quote(s string) string {
	return "\"" + s + "\""
}
