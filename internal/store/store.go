// Package store wraps the per-repo SQLite index database.
//
// The DB lives at ~/.cogni/<repo-hash>/index.db. v0.1 schema is embedded as
// schema.sql and applied on Open. Writers must serialize themselves; readers
// are concurrent-safe via SQLite's WAL mode.
package store

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// SchemaVersion is the current store schema. Bumped on incompatible changes.
const SchemaVersion = "1"

// Store owns a sql.DB handle to the index database.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite index at path and applies the schema.
// The caller must Close() when done.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the underlying DB handle.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the raw handle for advanced callers (tests, future tools).
func (s *Store) DB() *sql.DB { return s.db }

// migrate applies schema.sql and stamps schema_version in meta.
func (s *Store) migrate() error {
	if _, err := s.db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	_, err := s.db.Exec(
		`INSERT INTO meta(key, value) VALUES('schema_version', ?)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
		SchemaVersion,
	)
	if err != nil {
		return fmt.Errorf("stamp schema_version: %w", err)
	}
	return nil
}

// SchemaVersionFromDB returns the schema version recorded in the meta table.
func (s *Store) SchemaVersionFromDB() (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&v)
	return v, err
}
