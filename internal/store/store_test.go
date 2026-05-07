package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/islamborghini/cogni/internal/parse"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "index.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpenMigrate(t *testing.T) {
	s := openTemp(t)
	v, err := s.SchemaVersionFromDB()
	if err != nil {
		t.Fatalf("SchemaVersionFromDB: %v", err)
	}
	if v != SchemaVersion {
		t.Errorf("schema_version = %q, want %q", v, SchemaVersion)
	}
}

func TestUpsertAndReplaceSymbols(t *testing.T) {
	s := openTemp(t)

	src, err := os.ReadFile("../parse/testdata/simple.py")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	syms, err := parse.ParsePython(src)
	if err != nil {
		t.Fatalf("ParsePython: %v", err)
	}

	fid, err := s.UpsertFile(FileRow{
		Path: "simple.py", SHA256: "deadbeef",
		MtimeNs: 1, SizeBytes: int64(len(src)),
		LineCount: 25, Language: "python",
	})
	if err != nil {
		t.Fatalf("UpsertFile: %v", err)
	}
	if err := s.ReplaceSymbols(fid, syms); err != nil {
		t.Fatalf("ReplaceSymbols: %v", err)
	}

	got, err := s.SymbolsByName("hello", 10)
	if err != nil {
		t.Fatalf("SymbolsByName: %v", err)
	}
	if len(got) != 1 || got[0].Qualified != "hello" || got[0].StartLine != 6 {
		t.Errorf("hello lookup: got %+v", got)
	}

	// FTS5 trigram should match a substring of 'Greeter'.
	fts, err := s.SymbolsFTS("greet", "any", 10)
	if err != nil {
		t.Fatalf("SymbolsFTS: %v", err)
	}
	if len(fts) < 2 {
		t.Errorf("expected >=2 trigram matches for 'greet', got %d: %+v", len(fts), fts)
	}

	// Idempotency: replacing again with the same symbols should not duplicate.
	if err := s.ReplaceSymbols(fid, syms); err != nil {
		t.Fatalf("ReplaceSymbols again: %v", err)
	}
	again, _ := s.SymbolsByName("hello", 10)
	if len(again) != 1 {
		t.Errorf("after re-replace, expected 1 hello, got %d", len(again))
	}
}

func TestUpsertFileTwiceUpdatesRow(t *testing.T) {
	s := openTemp(t)
	id1, err := s.UpsertFile(FileRow{Path: "a.py", SHA256: "h1", MtimeNs: 1, SizeBytes: 10, LineCount: 1, Language: "python"})
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.UpsertFile(FileRow{Path: "a.py", SHA256: "h2", MtimeNs: 2, SizeBytes: 20, LineCount: 2, Language: "python"})
	if err != nil {
		t.Fatal(err)
	}
	if id1 != id2 {
		t.Errorf("upsert should keep id stable; got %d then %d", id1, id2)
	}
}
