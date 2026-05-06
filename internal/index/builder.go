// Package index walks a repo and writes parsed symbols into the store.
package index

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/islamborghini/cogni/internal/parse"
	"github.com/islamborghini/cogni/internal/store"
)

// Stats summarizes a build run.
type Stats struct {
	FilesScanned int
	FilesIndexed int
	FilesSkipped int
	Symbols      int
	Duration     time.Duration
	Errors       []error
}

// skipDirs are directory names we never descend into.
var skipDirs = map[string]bool{
	".git":         true,
	"__pycache__":  true,
	".venv":        true,
	"venv":         true,
	".tox":         true,
	"node_modules": true,
	".mypy_cache":  true,
	".pytest_cache": true,
	"dist":         true,
	"build":        true,
}

// Build walks root and indexes every Python file into s. Returns stats and
// the first fatal walk error, if any. Per-file errors are recorded in
// stats.Errors and do not abort the walk.
func Build(root string, s *store.Store) (Stats, error) {
	stats := Stats{}
	start := time.Now()

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Errorf("%s: %w", path, err))
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".py") {
			return nil
		}
		stats.FilesScanned++
		n, err := indexFile(root, path, s)
		if err != nil {
			stats.FilesSkipped++
			stats.Errors = append(stats.Errors, fmt.Errorf("%s: %w", path, err))
			return nil
		}
		stats.FilesIndexed++
		stats.Symbols += n
		return nil
	})

	stats.Duration = time.Since(start)
	return stats, walkErr
}

// indexFile reads, hashes, parses, and writes one file. Returns symbol count.
func indexFile(root, path string, s *store.Store) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	sum := sha256.Sum256(src)
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)

	syms, err := parse.ParsePython(src)
	if err != nil {
		return 0, fmt.Errorf("parse: %w", err)
	}

	fid, err := s.UpsertFile(store.FileRow{
		Path:      rel,
		SHA256:    hex.EncodeToString(sum[:]),
		MtimeNs:   info.ModTime().UnixNano(),
		SizeBytes: info.Size(),
		LineCount: countLines(src),
		Language:  "python",
	})
	if err != nil {
		return 0, fmt.Errorf("upsert file: %w", err)
	}
	if err := s.ReplaceSymbols(fid, syms); err != nil {
		return 0, fmt.Errorf("replace symbols: %w", err)
	}
	return len(syms), nil
}

func countLines(src []byte) int {
	if len(src) == 0 {
		return 0
	}
	n := bytes.Count(src, []byte{'\n'})
	if src[len(src)-1] != '\n' {
		n++
	}
	return n
}

// discardLogger is a no-op writer used when callers don't want logs.
var _ io.Writer = io.Discard
