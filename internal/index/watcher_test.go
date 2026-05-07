package index

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/islamborghini/cogni/internal/store"
)

func TestWatcherReindexesOnWrite(t *testing.T) {
	root := t.TempDir()
	pyPath := filepath.Join(root, "a.py")
	if err := os.WriteFile(pyPath, []byte("def hello():\n    return 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := store.Open(filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := Build(root, s); err != nil {
		t.Fatal(err)
	}

	w, err := NewWatcher(root, s)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	w.debounce = 50 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = w.Run(ctx) }()

	if err := os.WriteFile(pyPath, []byte("def hello():\n    return 1\n\ndef added():\n    return 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait briefly for the event, then flush the debounce window.
	time.Sleep(100 * time.Millisecond)
	w.Flush()

	got, err := s.SymbolsByName("added", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 'added' to be indexed after write, got %d hits", len(got))
	}
}

func TestWatcherShortCircuitsOnUnchangedContent(t *testing.T) {
	root := t.TempDir()
	pyPath := filepath.Join(root, "a.py")
	original := []byte("def hello():\n    return 1\n")
	if err := os.WriteFile(pyPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := Build(root, s); err != nil {
		t.Fatal(err)
	}

	// Capture the original sha and last_indexed by reading directly.
	beforeSHA, err := s.FileSHA("a.py")
	if err != nil {
		t.Fatal(err)
	}

	w, err := NewWatcher(root, s)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	w.debounce = 30 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = w.Run(ctx) }()

	// Re-write identical bytes. The sha won't change → short-circuit hits.
	if err := os.WriteFile(pyPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(80 * time.Millisecond)
	w.Flush()

	afterSHA, err := s.FileSHA("a.py")
	if err != nil {
		t.Fatal(err)
	}
	if beforeSHA != afterSHA {
		t.Errorf("sha mismatch unexpectedly: %s -> %s", beforeSHA, afterSHA)
	}
}

func TestWatcherDeletesRowOnRemove(t *testing.T) {
	root := t.TempDir()
	pyPath := filepath.Join(root, "a.py")
	if err := os.WriteFile(pyPath, []byte("def hello():\n    return 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := Build(root, s); err != nil {
		t.Fatal(err)
	}

	w, err := NewWatcher(root, s)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	w.debounce = 30 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = w.Run(ctx) }()

	if err := os.Remove(pyPath); err != nil {
		t.Fatal(err)
	}
	time.Sleep(80 * time.Millisecond)
	w.Flush()

	if _, err := s.FileSHA("a.py"); err == nil {
		t.Errorf("expected error after remove, file row still present")
	}
}
