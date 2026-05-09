package bench

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriteReportContainsHeaderAndRows(t *testing.T) {
	meta := ReportMeta{
		TargetRepo:       "https://github.com/encode/httpx",
		TargetSHA:        "abc1234",
		RunsPerCondition: 5,
		GeneratedAt:      time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC),
		CogniVersion:     "v0.1.0",
	}
	summaries := []TaskSummary{
		{
			TaskID:            "explain-transports",
			Family:            FamilyExplain,
			Baseline:          CondStats{Runs: 5, PassCount: 5, MeanTotal: 1200},
			Cogni:             CondStats{Runs: 5, PassCount: 5, MeanTotal: 800},
			TokenReductionPct: 33.3,
		},
	}
	var buf bytes.Buffer
	if err := WriteReport(&buf, meta, summaries); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	wantSubs := []string{
		"# Cogni benchmark report",
		"https://github.com/encode/httpx",
		"`abc1234`",
		"Runs per condition:** 5",
		"v0.1.0",
		"2026-05-07T12:00:00Z",
		"+33.3%",
		"explain-transports",
		"explain",
		"5/5",
		"BENCHMARK.md",
	}
	for _, s := range wantSubs {
		if !strings.Contains(out, s) {
			t.Errorf("output missing %q\n--- output ---\n%s", s, out)
		}
	}
}

func TestWriteReportEmpty(t *testing.T) {
	var buf bytes.Buffer
	err := WriteReport(&buf, ReportMeta{GeneratedAt: time.Now()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No tasks were scored") {
		t.Errorf("expected empty-state message, got:\n%s", buf.String())
	}
}

func TestWriteReportOverallMeanIsAverageOfReductions(t *testing.T) {
	summaries := []TaskSummary{
		{TaskID: "a", TokenReductionPct: 20},
		{TaskID: "b", TokenReductionPct: 40},
	}
	var buf bytes.Buffer
	_ = WriteReport(&buf, ReportMeta{GeneratedAt: time.Now()}, summaries)
	if !strings.Contains(buf.String(), "+30.0%") {
		t.Errorf("expected mean of 30%% in output:\n%s", buf.String())
	}
}

func TestWriteReportIncludesFailureDetails(t *testing.T) {
	summaries := []TaskSummary{
		{
			TaskID:   "bug",
			Family:   FamilyBugFix,
			Baseline: CondStats{Runs: 1, PassCount: 0},
			Cogni:    CondStats{Runs: 1, PassCount: 1},
		},
	}
	scores := []Score{
		{
			Run: RunResult{TaskID: "bug", Condition: ConditionBaseline, RunIndex: 0},
			Criteria: []CriterionResult{
				{
					Criterion: Criterion{TestsPass: "tests/test_bug.py::test_case"},
					Status:    StatusFail,
					Detail:    "pytest failed: exit status 1",
				},
			},
		},
	}
	var buf bytes.Buffer
	if err := WriteReport(&buf, ReportMeta{GeneratedAt: time.Now()}, summaries, scores...); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"## Failed runs", "`bug/baseline/0`", "tests_pass=", "pytest failed"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n--- output ---\n%s", want, out)
		}
	}
}
