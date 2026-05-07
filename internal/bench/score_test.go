package bench

import (
	"errors"
	"strings"
	"testing"
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

func TestScoreRunSkipsUnimplementedCriteria(t *testing.T) {
	task := Task{
		ID:     "t1",
		Family: FamilyBugFix,
		Prompt: "p",
		SuccessCriteria: []Criterion{
			{TestsPass: "tests/foo"},
			{FunctionAdded: "httpx.bar"},
		},
	}
	s := ScoreRun(task, RunResult{})
	// All criteria skipped → no checks → Pass should be false (nothing
	// was actually verified).
	if s.Pass {
		t.Errorf("all-skipped run should not pass")
	}
	if s.NumChecked != 0 {
		t.Errorf("checked=%d, want 0", s.NumChecked)
	}
	for _, cr := range s.Criteria {
		if cr.Status != StatusSkipped {
			t.Errorf("criterion %+v should be skipped, got %s", cr.Criterion, cr.Status)
		}
	}
}

func TestScoreRunMixedPassAndSkip(t *testing.T) {
	task := Task{
		ID:     "t1",
		Family: FamilyExplain,
		Prompt: "p",
		SuccessCriteria: []Criterion{
			{OutputContains: "ok"},
			{TestsPass: "tests/foo"},
		},
	}
	s := ScoreRun(task, RunResult{Output: "all ok"})
	if !s.Pass {
		t.Errorf("expected pass: one passed, one skipped")
	}
	if s.NumChecked != 1 || s.NumPassed != 1 {
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
