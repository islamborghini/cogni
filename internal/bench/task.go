// Package bench defines the benchmark task format and runner used by
// `cogni bench`. The methodology is described in BENCHMARK.md at the repo
// root; this package is the executable counterpart.
package bench

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Family classifies a task into one of the four families defined in
// BENCHMARK.md (#task-selection-criteria). The values are stable strings so
// they can be filtered from the CLI and emitted in reports.
type Family string

const (
	FamilyExplain    Family = "explain"
	FamilyBugFix     Family = "bug-fix"
	FamilyAddFeature Family = "add-feature"
	FamilyRefactor   Family = "refactor"
)

// Criterion is one checkable assertion against an agent run. Exactly one of
// the fields is non-empty for a given criterion.
type Criterion struct {
	OutputContains string `yaml:"output_contains,omitempty"`
	FileModified   string `yaml:"file_modified,omitempty"`
	TestsPass      string `yaml:"tests_pass,omitempty"`
	FunctionAdded  string `yaml:"function_added,omitempty"`
}

// Task is one benchmark task. The same Task is run twice per agent (baseline
// vs with-Cogni) for n runs each.
type Task struct {
	ID              string      `yaml:"id"`
	Family          Family      `yaml:"family"`
	Prompt          string      `yaml:"prompt"`
	Description     string      `yaml:"description,omitempty"`
	SuccessCriteria []Criterion `yaml:"success_criteria"`
	SourcePR        string      `yaml:"source_pr,omitempty"`
	SourceIssue     string      `yaml:"source_issue,omitempty"`
}

// TaskSet is the YAML root: a list of tasks plus shared metadata.
type TaskSet struct {
	TargetRepo string `yaml:"target_repo"`
	TargetSHA  string `yaml:"target_sha"`
	Tasks      []Task `yaml:"tasks"`
}

// LoadTasks parses a task definition file. Returns a fully-populated TaskSet
// or an error describing the first invalid task.
func LoadTasks(path string) (*TaskSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}
	var set TaskSet
	if err := yaml.Unmarshal(data, &set); err != nil {
		return nil, fmt.Errorf("parse tasks: %w", err)
	}
	if err := set.Validate(); err != nil {
		return nil, err
	}
	return &set, nil
}

// Validate checks structural invariants enforced by the methodology: every
// task has an id, family, prompt, and at least one success criterion; ids
// are unique.
func (ts *TaskSet) Validate() error {
	seen := map[string]bool{}
	for i, t := range ts.Tasks {
		if t.ID == "" {
			return fmt.Errorf("task[%d]: id is required", i)
		}
		if seen[t.ID] {
			return fmt.Errorf("task[%d] %q: duplicate id", i, t.ID)
		}
		seen[t.ID] = true
		switch t.Family {
		case FamilyExplain, FamilyBugFix, FamilyAddFeature, FamilyRefactor:
		default:
			return fmt.Errorf("task %q: unknown family %q", t.ID, t.Family)
		}
		if t.Prompt == "" {
			return fmt.Errorf("task %q: prompt is required", t.ID)
		}
		if len(t.SuccessCriteria) == 0 {
			return fmt.Errorf("task %q: at least one success_criteria entry is required", t.ID)
		}
		for j, c := range t.SuccessCriteria {
			if c.OutputContains == "" && c.FileModified == "" &&
				c.TestsPass == "" && c.FunctionAdded == "" {
				return fmt.Errorf("task %q: success_criteria[%d] is empty", t.ID, j)
			}
		}
	}
	return nil
}
