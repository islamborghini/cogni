package bench

import (
	"math"
	"testing"
)

func score(taskID string, cond Condition, in, out int, pass bool) Score {
	return Score{
		Run:  RunResult{TaskID: taskID, Condition: cond, InputTokens: in, OutputTokens: out},
		Pass: pass,
	}
}

func TestAggregateBasicDelta(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "t1", Family: FamilyExplain}}}
	scores := []Score{
		score("t1", ConditionBaseline, 1000, 100, true),
		score("t1", ConditionBaseline, 1200, 100, true),
		score("t1", ConditionCogni, 600, 100, true),
		score("t1", ConditionCogni, 800, 100, false),
	}
	got := Aggregate(set, scores)
	if len(got) != 1 {
		t.Fatalf("want 1 summary, got %d", len(got))
	}
	s := got[0]
	if s.Baseline.Runs != 2 || s.Cogni.Runs != 2 {
		t.Errorf("runs: baseline=%d cogni=%d", s.Baseline.Runs, s.Cogni.Runs)
	}
	if s.Baseline.PassCount != 2 || s.Cogni.PassCount != 1 {
		t.Errorf("pass: baseline=%d cogni=%d", s.Baseline.PassCount, s.Cogni.PassCount)
	}
	// baseline total mean = (1100+1300)/2 = 1200; cogni = (700+900)/2 = 800
	// reduction = (1200-800)/1200 = 33.33%
	if math.Abs(s.TokenReductionPct-33.333333) > 0.01 {
		t.Errorf("reduction = %f, want ~33.33", s.TokenReductionPct)
	}
}

func TestAggregateNegativeReduction(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "t1"}}}
	scores := []Score{
		score("t1", ConditionBaseline, 100, 0, true),
		score("t1", ConditionCogni, 200, 0, true),
	}
	got := Aggregate(set, scores)
	if got[0].TokenReductionPct >= 0 {
		t.Errorf("expected negative reduction, got %f", got[0].TokenReductionPct)
	}
}

func TestAggregateMedianOddEven(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "t1"}}}
	scores := []Score{
		score("t1", ConditionBaseline, 10, 0, true),
		score("t1", ConditionBaseline, 30, 0, true),
		score("t1", ConditionBaseline, 20, 0, true),
		score("t1", ConditionCogni, 100, 0, true),
		score("t1", ConditionCogni, 200, 0, true),
	}
	got := Aggregate(set, scores)
	if got[0].Baseline.MedianTotal != 20 {
		t.Errorf("odd median = %d, want 20", got[0].Baseline.MedianTotal)
	}
	if got[0].Cogni.MedianTotal != 150 {
		t.Errorf("even median = %d, want 150", got[0].Cogni.MedianTotal)
	}
}

func TestAggregateOmitsTasksWithoutScores(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "a"}, {ID: "b"}}}
	got := Aggregate(set, []Score{score("b", ConditionBaseline, 10, 0, true)})
	if len(got) != 1 || got[0].TaskID != "b" {
		t.Errorf("want only b, got %+v", got)
	}
}

func TestAggregateCostWeightedReduction(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "t1"}}}
	// Baseline: input=100, output=100, cache_create=0, cache_read=0
	//   weighted = 100 + 5*100 = 600
	// Cogni:    input=100, output=100, cache_create=200, cache_read=10000
	//   weighted = 100 + 5*100 + 1.25*200 + 0.1*10000 = 1850
	// Raw token reduction = 0 (both 200). Cost reduction = (600-1850)/600 ≈ -208.3%.
	scores := []Score{
		{Run: RunResult{TaskID: "t1", Condition: ConditionBaseline, InputTokens: 100, OutputTokens: 100}, Pass: true},
		{Run: RunResult{TaskID: "t1", Condition: ConditionCogni, InputTokens: 100, OutputTokens: 100, CacheCreationTokens: 200, CacheReadTokens: 10000}, Pass: true},
	}
	got := Aggregate(set, scores)[0]
	if got.TokenReductionPct != 0 {
		t.Errorf("raw reduction = %f, want 0", got.TokenReductionPct)
	}
	if math.Abs(got.CostReductionPct-(-208.333333)) > 0.01 {
		t.Errorf("cost reduction = %f, want ~-208.33", got.CostReductionPct)
	}
	if math.Abs(got.Cogni.MeanCacheRead-10000) > 0.01 {
		t.Errorf("mean cache_read = %f, want 10000", got.Cogni.MeanCacheRead)
	}
}

func TestAggregateSortsByTaskID(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "z"}, {ID: "a"}, {ID: "m"}}}
	scores := []Score{
		score("z", ConditionBaseline, 1, 0, true),
		score("a", ConditionBaseline, 1, 0, true),
		score("m", ConditionBaseline, 1, 0, true),
	}
	got := Aggregate(set, scores)
	want := []string{"a", "m", "z"}
	for i, ts := range got {
		if ts.TaskID != want[i] {
			t.Errorf("pos %d = %s, want %s", i, ts.TaskID, want[i])
		}
	}
}
