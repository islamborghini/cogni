package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureRepo creates a tiny git repo with two commits and returns its path
// plus the SHA of the first commit (so checkouts can target a known state).
func fixtureRepo(t *testing.T) (path string, firstSHA string) {
	t.Helper()
	dir := t.TempDir()
	mustGit(t, dir, "init", "--quiet", "--initial-branch=main")
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "test")

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", "a.txt")
	mustGit(t, dir, "commit", "--quiet", "-m", "first")

	out := mustCaptureGit(t, dir, "rev-parse", "HEAD")
	first := strings.TrimSpace(out)

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "commit", "--quiet", "-am", "second")
	return dir, first
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if err := runGit(dir, args...); err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
}

func mustCaptureGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := captureGit(dir, args...)
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func TestPrepareChecksOutPinnedSHA(t *testing.T) {
	src, firstSHA := fixtureRepo(t)
	parent := t.TempDir()

	ws, err := Prepare("file://"+src, firstSHA, parent)
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	if !ws.Exists() {
		t.Fatal("workspace directory missing")
	}
	got, err := os.ReadFile(filepath.Join(ws.Path, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v1\n" {
		t.Errorf("a.txt = %q, want %q (wrong commit checked out)", got, "v1\n")
	}
}

func TestModifiedFilesDetectsEdits(t *testing.T) {
	src, sha := fixtureRepo(t)
	parent := t.TempDir()
	ws, err := Prepare("file://"+src, sha, parent)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ws.Cleanup() }()

	// modify a tracked file and add a new one
	if err := os.WriteFile(filepath.Join(ws.Path, "a.txt"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws.Path, "new.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mods, err := ws.ModifiedFiles()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"a.txt": false, "new.txt": false}
	for _, f := range mods {
		if _, ok := want[f]; ok {
			want[f] = true
		}
	}
	for f, seen := range want {
		if !seen {
			t.Errorf("expected %s in modified files, got %v", f, mods)
		}
	}
}

func TestDiffCapturesTrackedEdits(t *testing.T) {
	src, sha := fixtureRepo(t)
	ws, err := Prepare("file://"+src, sha, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ws.Cleanup() }()

	if err := os.WriteFile(filepath.Join(ws.Path, "a.txt"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	diff, err := ws.Diff()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diff, "+edited") {
		t.Fatalf("diff did not include edit:\n%s", diff)
	}
}

func TestPrepareRejectsUnpinnedSHA(t *testing.T) {
	cases := []string{"", "TBD-PINNED-BEFORE-MEASUREMENT"}
	for _, sha := range cases {
		_, err := Prepare("file:///nope", sha, t.TempDir())
		if err == nil {
			t.Errorf("sha=%q: expected error, got nil", sha)
		}
	}
}

func TestCleanupRemovesDir(t *testing.T) {
	src, sha := fixtureRepo(t)
	ws, err := Prepare("file://"+src, sha, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := ws.Path
	if err := ws.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("workspace dir still exists: %v", err)
	}
}
