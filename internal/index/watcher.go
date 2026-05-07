package index

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/islamborghini/cogni/internal/store"
)

// DefaultDebounce is the window we wait after the last event before reindexing.
// Editors often emit several rapid Write/Create events for a single save.
const DefaultDebounce = 250 * time.Millisecond

// Watcher reindexes Python files as they change on disk.
type Watcher struct {
	root     string
	store    *store.Store
	fs       *fsnotify.Watcher
	debounce time.Duration

	mu      sync.Mutex
	pending map[string]struct{}
	timer   *time.Timer
}

// NewWatcher creates a watcher rooted at root and using s as the index store.
// Use Run to start the event loop.
func NewWatcher(root string, s *store.Store) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		root:     root,
		store:    s,
		fs:       fw,
		debounce: DefaultDebounce,
		pending:  make(map[string]struct{}),
	}
	if err := w.addTree(root); err != nil {
		_ = fw.Close()
		return nil, err
	}
	return w, nil
}

// Close releases the underlying fsnotify resources.
func (w *Watcher) Close() error { return w.fs.Close() }

// Flush forces any pending debounced reindex to run synchronously. Tests use
// this to avoid sleeping on the debounce window.
func (w *Watcher) Flush() { w.fire() }

// Run blocks until ctx is canceled, processing fsnotify events.
func (w *Watcher) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-w.fs.Events:
			if !ok {
				return nil
			}
			w.handle(ev)
		case err, ok := <-w.fs.Errors:
			if !ok {
				return nil
			}
			if err != nil && !errors.Is(err, fsnotify.ErrEventOverflow) {
				// Surface unexpected errors but keep running.
				_ = err
			}
		}
	}
}

func (w *Watcher) addTree(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		return w.fs.Add(path)
	})
}

func (w *Watcher) handle(ev fsnotify.Event) {
	// New directories should also be watched.
	if ev.Op&fsnotify.Create != 0 {
		if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
			if !skipDirs[filepath.Base(ev.Name)] {
				_ = w.addTree(ev.Name)
			}
			return
		}
	}

	if !strings.HasSuffix(ev.Name, ".py") {
		return
	}

	if ev.Op&fsnotify.Remove != 0 || ev.Op&fsnotify.Rename != 0 {
		w.deleteFile(ev.Name)
		return
	}
	if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return
	}
	w.enqueue(ev.Name)
}

func (w *Watcher) enqueue(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending[path] = struct{}{}
	if w.timer == nil {
		w.timer = time.AfterFunc(w.debounce, w.fire)
	} else {
		w.timer.Reset(w.debounce)
	}
}

func (w *Watcher) fire() {
	w.mu.Lock()
	paths := w.pending
	w.pending = make(map[string]struct{})
	w.timer = nil
	w.mu.Unlock()

	for p := range paths {
		w.reindex(p)
	}
}

// reindex re-parses path if its content hash differs from the stored one.
func (w *Watcher) reindex(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		// File was removed before we got to it.
		w.deleteFile(path)
		return
	}
	rel := w.rel(path)
	sum := sha256.Sum256(src)
	hashHex := hex.EncodeToString(sum[:])

	if existing, err := w.store.FileSHA(rel); err == nil && existing == hashHex {
		return
	}
	_, _, _ = indexFile(w.root, path, w.store)
}

func (w *Watcher) deleteFile(path string) {
	rel := w.rel(path)
	_ = w.store.DeleteFile(rel)
}

func (w *Watcher) rel(path string) string {
	rel, err := filepath.Rel(w.root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
