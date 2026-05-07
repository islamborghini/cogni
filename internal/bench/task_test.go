package bench

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTasksYAMLValid(t *testing.T) {
	// The shipped tasks.yaml must always parse and validate, since CI relies
	// on it being loadable even while individual tasks are placeholders.
	path := filepath.Join("tasks.yaml")
	set, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(set.Tasks) == 0 {
		t.Fatal("expected at least one task")
	}
	if set.TargetRepo == "" {
		t.Error("target_repo missing")
	}
}

func TestValidateRejectsBad(t *testing.T) {
	cases := []struct {
		name string
		set  TaskSet
		want string
	}{
		{
			"missing id",
			TaskSet{Tasks: []Task{{Family: FamilyExplain, Prompt: "p", SuccessCriteria: []Criterion{{OutputContains: "x"}}}}},
			"id is required",
		},
		{
			"unknown family",
			TaskSet{Tasks: []Task{{ID: "a", Family: "weird", Prompt: "p", SuccessCriteria: []Criterion{{OutputContains: "x"}}}}},
			"unknown family",
		},
		{
			"missing prompt",
			TaskSet{Tasks: []Task{{ID: "a", Family: FamilyExplain, SuccessCriteria: []Criterion{{OutputContains: "x"}}}}},
			"prompt is required",
		},
		{
			"no criteria",
			TaskSet{Tasks: []Task{{ID: "a", Family: FamilyExplain, Prompt: "p"}}},
			"at least one success_criteria",
		},
		{
			"empty criterion",
			TaskSet{Tasks: []Task{{ID: "a", Family: FamilyExplain, Prompt: "p", SuccessCriteria: []Criterion{{}}}}},
			"is empty",
		},
		{
			"duplicate id",
			TaskSet{Tasks: []Task{
				{ID: "a", Family: FamilyExplain, Prompt: "p", SuccessCriteria: []Criterion{{OutputContains: "x"}}},
				{ID: "a", Family: FamilyExplain, Prompt: "p", SuccessCriteria: []Criterion{{OutputContains: "x"}}},
			}},
			"duplicate id",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.set.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want substring %q", err, tc.want)
			}
		})
	}
}
