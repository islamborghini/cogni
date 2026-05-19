package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureResumeMetaWritesOnFirstUse(t *testing.T) {
	dir := t.TempDir()
	set := &TaskSet{
		TargetRepo: "https://example/repo",
		TargetSHA:  "abc",
		Tasks:      []Task{{ID: "t1", Prompt: "p"}},
	}
	opts := HarnessOptions{RunnerCfg: ClaudeRunnerConfig{RunsDir: dir, Model: "claude-sonnet-4-6"}}
	path, err := ensureResumeMeta(set, opts)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, resumeMetaFilename) {
		t.Errorf("path=%q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("meta file not written: %v", err)
	}
}

func TestEnsureResumeMetaAcceptsMatchingSecondCall(t *testing.T) {
	dir := t.TempDir()
	set := &TaskSet{TargetSHA: "abc", Tasks: []Task{{ID: "t", Prompt: "p"}}}
	opts := HarnessOptions{RunnerCfg: ClaudeRunnerConfig{RunsDir: dir, Model: "m"}}
	if _, err := ensureResumeMeta(set, opts); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureResumeMeta(set, opts); err != nil {
		t.Errorf("matching second call should succeed, got %v", err)
	}
}

func TestEnsureResumeMetaRejectsChangedTask(t *testing.T) {
	dir := t.TempDir()
	set := &TaskSet{TargetSHA: "abc", Tasks: []Task{{ID: "t", Prompt: "original"}}}
	opts := HarnessOptions{RunnerCfg: ClaudeRunnerConfig{RunsDir: dir, Model: "m"}}
	if _, err := ensureResumeMeta(set, opts); err != nil {
		t.Fatal(err)
	}
	set.Tasks[0].Prompt = "edited"
	_, err := ensureResumeMeta(set, opts)
	if err == nil || !strings.Contains(err.Error(), "tasks") {
		t.Errorf("expected tasks mismatch, got %v", err)
	}
}

func TestEnsureResumeMetaRejectsChangedModel(t *testing.T) {
	dir := t.TempDir()
	set := &TaskSet{TargetSHA: "abc", Tasks: []Task{{ID: "t", Prompt: "p"}}}
	opts := HarnessOptions{RunnerCfg: ClaudeRunnerConfig{RunsDir: dir, Model: "m1"}}
	if _, err := ensureResumeMeta(set, opts); err != nil {
		t.Fatal(err)
	}
	opts.RunnerCfg.Model = "m2"
	_, err := ensureResumeMeta(set, opts)
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Errorf("expected model mismatch, got %v", err)
	}
}

func TestEnsureResumeMetaSkipsWhenNoRunsDir(t *testing.T) {
	set := &TaskSet{Tasks: []Task{{ID: "t"}}}
	path, err := ensureResumeMeta(set, HarnessOptions{})
	if err != nil || path != "" {
		t.Errorf("no RunsDir → expected ('', nil), got (%q, %v)", path, err)
	}
}
