package bench

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestRunLoopRunsAllTasksConditionsAndRuns(t *testing.T) {
	set := &TaskSet{Tasks: []Task{
		{ID: "a", Family: FamilyExplain, Prompt: "p", SuccessCriteria: []Criterion{{OutputContains: "x"}}},
		{ID: "b", Family: FamilyExplain, Prompt: "p", SuccessCriteria: []Criterion{{OutputContains: "x"}}},
	}}
	opts := HarnessOptions{RunsPerCondition: 3}
	var seen []string
	fake := func(ctx context.Context, task Task, cond Condition, idx int) Score {
		seen = append(seen, fmt.Sprintf("%s/%s/%d", task.ID, cond, idx))
		return Score{Run: RunResult{TaskID: task.ID, Condition: cond, RunIndex: idx}}
	}
	scores, err := runLoop(context.Background(), set, opts, fake)
	if err != nil {
		t.Fatal(err)
	}
	want := 2 * 2 * 3
	if len(scores) != want {
		t.Errorf("scores=%d, want %d", len(scores), want)
	}
	if len(seen) != want {
		t.Fatalf("calls=%d, want %d", len(seen), want)
	}
	// First six calls should be a/baseline/0..2 then a/cogni/0..2
	wantPrefix := []string{"a/baseline/0", "a/baseline/1", "a/baseline/2", "a/cogni/0", "a/cogni/1", "a/cogni/2"}
	for i, w := range wantPrefix {
		if seen[i] != w {
			t.Errorf("seen[%d]=%s, want %s", i, seen[i], w)
		}
	}
}

func TestRunLoopDefaultsApplied(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "a"}}}
	calls := 0
	fake := func(ctx context.Context, _ Task, _ Condition, _ int) Score {
		calls++
		return Score{}
	}
	if _, err := runLoop(context.Background(), set, HarnessOptions{}, fake); err != nil {
		t.Fatal(err)
	}
	// default n=5, 2 conditions, 1 task = 10 calls
	if calls != 10 {
		t.Errorf("calls=%d, want 10 (defaults: n=5, 2 conditions)", calls)
	}
}

func TestRunLoopHonoursContextCancel(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "a"}, {ID: "b"}}}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	fake := func(ctx context.Context, _ Task, _ Condition, _ int) Score {
		calls++
		if calls == 2 {
			cancel()
		}
		return Score{}
	}
	scores, err := runLoop(ctx, set, HarnessOptions{RunsPerCondition: 5}, fake)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err=%v, want context.Canceled", err)
	}
	if len(scores) == 0 {
		t.Error("expected partial scores before cancel")
	}
	if len(scores) >= 2*2*5 {
		t.Errorf("scores=%d, expected fewer than full count", len(scores))
	}
}

func TestRunLoopRejectsNilSet(t *testing.T) {
	_, err := runLoop(context.Background(), nil, HarnessOptions{}, nil)
	if err == nil {
		t.Error("expected error on nil task set")
	}
}
