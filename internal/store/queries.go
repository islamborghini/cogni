package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/islamborghini/cogni/internal/parse"
)

// FileRow describes a row in the files table.
type FileRow struct {
	ID         int64
	Path       string
	SHA256     string
	MtimeNs    int64
	SizeBytes  int64
	LineCount  int
	Language   string
	LastIndexed int64
}

// SymbolRow is the persisted form of parse.Symbol joined with its file path.
type SymbolRow struct {
	ID        int64
	FileID    int64
	FilePath  string
	Name      string
	Qualified string
	Kind      string
	StartLine int
	EndLine   int
	Signature string
	Docstring string
}

// UpsertFile inserts a file row or updates the existing row at the same path.
// Returns the file id.
func (s *Store) UpsertFile(f FileRow) (int64, error) {
	if f.LastIndexed == 0 {
		f.LastIndexed = time.Now().UnixNano()
	}
	res, err := s.db.Exec(
		`INSERT INTO files(path, sha256, mtime_ns, size_bytes, line_count, language, last_indexed)
		 VALUES(?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(path) DO UPDATE SET
		   sha256=excluded.sha256,
		   mtime_ns=excluded.mtime_ns,
		   size_bytes=excluded.size_bytes,
		   line_count=excluded.line_count,
		   language=excluded.language,
		   last_indexed=excluded.last_indexed`,
		f.Path, f.SHA256, f.MtimeNs, f.SizeBytes, f.LineCount, f.Language, f.LastIndexed,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert file: %w", err)
	}
	var id int64
	if err := s.db.QueryRow(`SELECT id FROM files WHERE path=?`, f.Path).Scan(&id); err != nil {
		return 0, fmt.Errorf("read file id: %w", err)
	}
	_ = res
	return id, nil
}

// ReplaceSymbols deletes the existing symbols for fileID and inserts the
// given symbols in a single transaction. The FTS5 mirror updates via
// triggers. Use this from the indexer after each (re)parse.
func (s *Store) ReplaceSymbols(fileID int64, syms []parse.Symbol) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM symbols WHERE file_id=?`, fileID); err != nil {
		return fmt.Errorf("clear symbols: %w", err)
	}
	stmt, err := tx.Prepare(
		`INSERT INTO symbols(file_id, name, qualified, kind, start_line, end_line, signature, docstring)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()
	for _, sym := range syms {
		if _, err := stmt.Exec(
			fileID, sym.Name, sym.Qualified, string(sym.Kind),
			sym.StartLine, sym.EndLine,
			nullableString(sym.Signature), nullableString(sym.Docstring),
		); err != nil {
			return fmt.Errorf("insert symbol %s: %w", sym.Qualified, err)
		}
	}
	return tx.Commit()
}

// SymbolsByName returns symbols whose bare name exactly matches q.
func (s *Store) SymbolsByName(q string, limit int) ([]SymbolRow, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.file_id, f.path, s.name, s.qualified, s.kind,
		        s.start_line, s.end_line,
		        IFNULL(s.signature, ''), IFNULL(s.docstring, '')
		   FROM symbols s
		   JOIN files f ON f.id = s.file_id
		  WHERE s.name = ?
		  ORDER BY f.path, s.start_line
		  LIMIT ?`,
		q, limit,
	)
	if err != nil {
		return nil, err
	}
	return scanSymbols(rows)
}

// SymbolsFTS runs a trigram FTS5 query over name/qualified/docstring and
// returns the matching symbol rows, ordered by relevance.
func (s *Store) SymbolsFTS(q string, limit int) ([]SymbolRow, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.file_id, f.path, s.name, s.qualified, s.kind,
		        s.start_line, s.end_line,
		        IFNULL(s.signature, ''), IFNULL(s.docstring, '')
		   FROM symbols_fts fts
		   JOIN symbols s ON s.id = fts.rowid
		   JOIN files   f ON f.id = s.file_id
		  WHERE symbols_fts MATCH ?
		  ORDER BY rank
		  LIMIT ?`,
		q, limit,
	)
	if err != nil {
		return nil, err
	}
	return scanSymbols(rows)
}

func scanSymbols(rows *sql.Rows) ([]SymbolRow, error) {
	defer rows.Close()
	var out []SymbolRow
	for rows.Next() {
		var r SymbolRow
		if err := rows.Scan(
			&r.ID, &r.FileID, &r.FilePath, &r.Name, &r.Qualified, &r.Kind,
			&r.StartLine, &r.EndLine, &r.Signature, &r.Docstring,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
