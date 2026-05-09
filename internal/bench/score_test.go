package bench

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScoreRunOutputContains(t *testing.T) {
	task := Task{
		ID:     "t1",
		Family: FamilyExplain,
		Prompt: "p",
		SuccessCriteria: []Criterion{
			{OutputContains: "HTTPTransport"},
		},
	}
	run := RunResult{Output: "uses HTTPTransport for sync calls"}
	s := ScoreRun(task, run)
	if !s.Pass {
		t.Fatalf("want pass, got fail: %+v", s.Criteria)
	}
	if s.NumChecked != 1 || s.NumPassed != 1 {
		t.Errorf("counts: checked=%d passed=%d", s.NumChecked, s.NumPassed)
	}
}

func TestScoreRunOutputMissing(t *testing.T) {
	task := Task{
		ID:              "t1",
		Family:          FamilyExplain,
		Prompt:          "p",
		SuccessCriteria: []Criterion{{OutputContains: "needle"}},
	}
	s := ScoreRun(task, RunResult{Output: "haystack"})
	if s.Pass {
		t.Fatal("want fail")
	}
	if len(s.Criteria) != 1 || s.Criteria[0].Status != StatusFail {
		t.Fatalf("expected one fail, got %+v", s.Criteria)
	}
	if !strings.Contains(s.Criteria[0].Detail, "needle") {
		t.Errorf("detail should mention missing string: %q", s.Criteria[0].Detail)
	}
}

func TestScoreRunFileModified(t *testing.T) {
	task := Task{
		ID:              "t1",
		Family:          FamilyBugFix,
		Prompt:          "p",
		SuccessCriteria: []Criterion{{FileModified: "httpx/_client.py"}},
	}
	hit := ScoreRun(task, RunResult{FilesModified: []string{"httpx/_client.py", "tests/x.py"}})
	if !hit.Pass {
		t.Errorf("want pass for matching file")
	}
	miss := ScoreRun(task, RunResult{FilesModified: []string{"tests/x.py"}})
	if miss.Pass {
		t.Errorf("want fail for non-matching file")
	}
}

func TestScoreRunFunctionAdded(t *testing.T) {
	task := Task{
		ID:     "t1",
		Family: FamilyAddFeature,
		Prompt: "p",
		SuccessCriteria: []Criterion{
			{FunctionAdded: "httpx._utils.normalize_header_key"},
		},
	}
	diff := `diff --git a/httpx/_utils.py b/httpx/_utils.py
index 123..456 100644
--- a/httpx/_utils.py
+++ b/httpx/_utils.py
@@ -1,0 +2,2 @@
+def normalize_header_key(value):
+    return value.lower()
`
	s := ScoreRun(task, RunResult{Diff: diff})
	if !s.Pass {
		t.Fatalf("want pass, got %+v", s.Criteria)
	}
}

func TestScoreRunFunctionAddedAsync(t *testing.T) {
	task := Task{
		ID:              "t1",
		Family:          FamilyAddFeature,
		Prompt:          "p",
		SuccessCriteria: []Criterion{{FunctionAdded: "httpx.fetch"}},
	}
	diff := "+async def fetch(url):\n+    return url\n"
	s := ScoreRun(task, RunResult{Diff: diff})
	if !s.Pass {
		t.Fatalf("want pass for async def, got %+v", s.Criteria)
	}
}

func TestScoreRunFunctionAddedMissing(t *testing.T) {
	task := Task{
		ID:              "t1",
		Family:          FamilyAddFeature,
		Prompt:          "p",
		SuccessCriteria: []Criterion{{FunctionAdded: "httpx.target"}},
	}
	s := ScoreRun(task, RunResult{Diff: "+def other():\n+    pass\n"})
	if s.Pass {
		t.Fatal("want fail")
	}
	if len(s.Criteria) != 1 || s.Criteria[0].Status != StatusFail {
		t.Fatalf("expected one fail, got %+v", s.Criteria)
	}
}

func TestScoreRunTestsPass(t *testing.T) {
	task := Task{
		ID:     "t1",
		Family: FamilyBugFix,
		Prompt: "p",
		SuccessCriteria: []Criterion{
			{TestsPass: "tests/foo"},
		},
	}
	orig := runTestsPass
	t.Cleanup(func() { runTestsPass = orig })
	var gotWorkspace, gotNode string
	runTestsPass = func(workspace, node string) (string, error) {
		gotWorkspace, gotNode = workspace, node
		return "ok", nil
	}

	s := ScoreRun(task, RunResult{WorkspacePath: "/tmp/repo"})
	if !s.Pass {
		t.Fatalf("want pass, got %+v", s.Criteria)
	}
	if gotWorkspace != "/tmp/repo" || gotNode != "tests/foo" {
		t.Fatalf("runner got workspace=%q node=%q", gotWorkspace, gotNode)
	}
}

func TestScoreRunTestsPassFailure(t *testing.T) {
	task := Task{
		ID:              "t1",
		Family:          FamilyBugFix,
		Prompt:          "p",
		SuccessCriteria: []Criterion{{TestsPass: "tests/foo"}},
	}
	orig := runTestsPass
	t.Cleanup(func() { runTestsPass = orig })
	runTestsPass = func(workspace, node string) (string, error) {
		return "assertion failed", errors.New("exit status 1")
	}

	s := ScoreRun(task, RunResult{WorkspacePath: "/tmp/repo"})
	if s.Pass {
		t.Fatal("want fail")
	}
	if s.NumChecked != 1 || s.NumPassed != 0 {
		t.Errorf("counts: checked=%d passed=%d", s.NumChecked, s.NumPassed)
	}
	if !strings.Contains(s.Criteria[0].Detail, "assertion failed") {
		t.Fatalf("detail did not include pytest output: %q", s.Criteria[0].Detail)
	}
}

func TestScoreRunTestsPassRequiresWorkspace(t *testing.T) {
	task := Task{
		ID:              "t1",
		Family:          FamilyBugFix,
		Prompt:          "p",
		SuccessCriteria: []Criterion{{TestsPass: "tests/foo"}},
	}
	s := ScoreRun(task, RunResult{})
	if s.Pass {
		t.Fatal("want fail")
	}
	if s.NumChecked != 1 {
		t.Errorf("checked=%d, want 1", s.NumChecked)
	}
}

func TestEnsureTestVenvInstallsWhenMarkerMissing(t *testing.T) {
	workspace := t.TempDir()
	venv := filepath.Join(workspace, ".cogni-test-venv")
	python := venvPython(venv)
	if err := os.MkdirAll(filepath.Dir(python), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(python, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	orig := runCommand
	t.Cleanup(func() { runCommand = orig })
	var calls []string
	runCommand = func(timeout time.Duration, dir, name string, args ...string) (string, error) {
		calls = append(calls, name+" "+strings.Join(args, " "))
		return "", nil
	}

	got, err := ensureTestVenv(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if got != python {
		t.Fatalf("python=%q, want %q", got, python)
	}
	if len(calls) != 1 || !strings.Contains(calls[0], "pip install") {
		t.Fatalf("calls=%v, want one pip install call", calls)
	}
	if _, err := os.Stat(filepath.Join(venv, ".cogni-deps-installed")); err != nil {
		t.Fatalf("marker missing: %v", err)
	}
}

func TestScoreRunMixedPassAndSkip(t *testing.T) {
	task := Task{
		ID:     "t1",
		Family: FamilyExplain,
		Prompt: "p",
		SuccessCriteria: []Criterion{
			{OutputContains: "ok"},
			{FunctionAdded: "httpx.added"},
		},
	}
	s := ScoreRun(task, RunResult{
		Output: "all ok",
		Diff:   "+def added():\n+    return True\n",
	})
	if !s.Pass {
		t.Errorf("expected pass: output and function_added passed")
	}
	if s.NumChecked != 2 || s.NumPassed != 2 {
		t.Errorf("counts: checked=%d passed=%d", s.NumChecked, s.NumPassed)
	}
}

func TestScoreRunRunnerErrorIsFail(t *testing.T) {
	task := Task{
		ID:              "t1",
		Family:          FamilyExplain,
		Prompt:          "p",
		SuccessCriteria: []Criterion{{OutputContains: "anything"}},
	}
	s := ScoreRun(task, RunResult{Err: errors.New("network down")})
	if s.Pass {
		t.Error("runner error must not pass")
	}
	if len(s.Criteria) != 0 {
		t.Errorf("no criteria should be checked on runner error, got %d", len(s.Criteria))
	}
}
