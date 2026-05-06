package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/islamborghini/cogni/internal/store"
)

func TestBuildSmallTree(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(rel, body string) {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("a.py", "def hello():\n    return 1\n")
	mustWrite("pkg/b.py", "class B:\n    def m(self):\n        return 2\n")
	mustWrite(".git/config", "ignored\n")
	mustWrite("__pycache__/x.py", "should_be_skipped = 1\n")

	s, err := store.Open(filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	stats, err := Build(root, s)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if stats.FilesIndexed != 2 {
		t.Errorf("FilesIndexed = %d, want 2 (errors=%v)", stats.FilesIndexed, stats.Errors)
	}
	if stats.Symbols < 3 {
		t.Errorf("Symbols = %d, want >=3", stats.Symbols)
	}

	hits, err := s.SymbolsByName("hello", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Errorf("hello lookup: got %d, want 1", len(hits))
	}
}
