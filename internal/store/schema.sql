PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS files (
    id            INTEGER PRIMARY KEY,
    path          TEXT NOT NULL UNIQUE,
    sha256        TEXT NOT NULL,
    mtime_ns      INTEGER NOT NULL,
    size_bytes    INTEGER NOT NULL,
    line_count    INTEGER NOT NULL,
    language      TEXT NOT NULL,
    last_indexed  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_files_language ON files(language);

CREATE TABLE IF NOT EXISTS symbols (
    id            INTEGER PRIMARY KEY,
    file_id       INTEGER NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    qualified     TEXT NOT NULL,
    kind          TEXT NOT NULL,
    start_line    INTEGER NOT NULL,
    end_line      INTEGER NOT NULL,
    signature     TEXT,
    docstring     TEXT
);
CREATE INDEX IF NOT EXISTS idx_symbols_name      ON symbols(name);
CREATE INDEX IF NOT EXISTS idx_symbols_qualified ON symbols(qualified);
CREATE INDEX IF NOT EXISTS idx_symbols_file      ON symbols(file_id);
CREATE INDEX IF NOT EXISTS idx_symbols_kind      ON symbols(kind);

CREATE VIRTUAL TABLE IF NOT EXISTS symbols_fts USING fts5(
    name, qualified, docstring,
    content='symbols', content_rowid='id',
    tokenize='trigram'
);

CREATE TRIGGER IF NOT EXISTS symbols_ai AFTER INSERT ON symbols BEGIN
    INSERT INTO symbols_fts(rowid, name, qualified, docstring)
    VALUES (new.id, new.name, new.qualified, new.docstring);
END;

CREATE TRIGGER IF NOT EXISTS symbols_ad AFTER DELETE ON symbols BEGIN
    INSERT INTO symbols_fts(symbols_fts, rowid, name, qualified, docstring)
    VALUES ('delete', old.id, old.name, old.qualified, old.docstring);
END;

CREATE TRIGGER IF NOT EXISTS symbols_au AFTER UPDATE ON symbols BEGIN
    INSERT INTO symbols_fts(symbols_fts, rowid, name, qualified, docstring)
    VALUES ('delete', old.id, old.name, old.qualified, old.docstring);
    INSERT INTO symbols_fts(rowid, name, qualified, docstring)
    VALUES (new.id, new.name, new.qualified, new.docstring);
END;
