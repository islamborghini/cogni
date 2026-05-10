# Cogni

[![Release](https://img.shields.io/github/v/release/islamborghini/cogni)](https://github.com/islamborghini/cogni/releases)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/islamborghini/cogni.svg)](https://pkg.go.dev/github.com/islamborghini/cogni)

Open-core MCP (Model Context Protocol) server that drops into AI coding agents (Claude Code, Cursor, Windsurf, OpenCode) and reduces token consumption by giving them structured recall over a repository.

AI coding agents waste a large share of their token budget orienting in unfamiliar repos: re-reading files, grepping for symbols, and accumulating stale context. Cogni indexes a Python repository into a local SQLite code graph and exposes scoped MCP tools so agents move from overview to file to symbol to source instead of brute-forcing context.

## Quickstart

```sh
brew tap islamborghini/tap
brew install islamborghini/tap/cogni

cd path/to/your/python/repo
cogni install
```

Restart Claude Code. That's it.

`cogni install` does three things: registers Cogni in your Claude Code MCP config, writes a `CLAUDE.md` so the agent prefers Cogni's scoped tools over Read/Grep/Glob, and indexes the repository. Safe to re-run.

## Benchmark results

Measured on the `httpx` codebase using Claude Code SDK with n=5 runs per condition:

| Task | Family | Baseline tokens | Cogni tokens | Reduction | Pass rate |
|---|---|---:|---:|---:|---:|
| `explain-transports` | explain | 2,209 | 1,673 | **24.3% lower** | 5/5 vs 5/5 |

Methodology, fixed inputs, and reproduction steps are versioned in [BENCHMARK.md](BENCHMARK.md). Raw report: [bench-report.md](bench-report.md).

## Other install methods

### Go

```sh
go install github.com/islamborghini/cogni/cmd/cogni@latest
```

Requires Go 1.26+ and a working CGO toolchain (tree-sitter and SQLite are linked natively).

### From source

```sh
git clone https://github.com/islamborghini/cogni
cd cogni
go build -o cogni ./cmd/cogni
```

## Manual configuration

If you do not want to use `cogni install`, configure things by hand.

**1. Register Cogni in your Claude Code MCP config** (`~/.claude.json` or per-project `.mcp.json`):

```json
{
  "mcpServers": {
    "cogni": {
      "command": "cogni",
      "args": ["serve", "--root", "/absolute/path/to/repo"]
    }
  }
}
```

**2. Drop a `CLAUDE.md` at the repository root** so Claude Code knows when to call Cogni's tools:

```markdown
# Repository tooling

This repository has the **cogni** MCP server registered with these tools:

- `mcp__cogni__repo_overview`: high-level package map. Call FIRST for any architecture / "how does X work" question.
- `mcp__cogni__file_outline`: list symbols in a file without returning source. Use INSTEAD OF Read when you only need to know what a file exposes.
- `mcp__cogni__symbol_search`: find where a symbol is defined. Use INSTEAD OF Grep for definition lookups.
- `mcp__cogni__symbol_source`: return one symbol's body. Use INSTEAD OF Read after symbol_search.
- `mcp__cogni__find_references`: find usages of a symbol. Use INSTEAD OF Grep when answering "where is X used".

For code exploration tasks, **prefer the cogni tools** over Glob / Grep / Read. They return structured, scoped results and use 5-10x fewer tokens for the same answer.
```

**3. Index the repo and restart Claude Code:**

```sh
cogni index .
```

## Verifying it works

Open Claude Code in a repo where you ran `cogni install` and ask something like:

```
use repo_overview to explain how this codebase is organized
```

The tool call panel should show `mcp__cogni__repo_overview` instead of a series of Read/Glob/Grep calls.

## Tools

Cogni exposes five MCP tools:

| Tool | Purpose |
|---|---|
| `repo_overview` | Package/module tree with file counts and top symbols. |
| `file_outline` | Symbols in one file with line ranges, signatures, and docstrings. |
| `symbol_search` | Indexed definition search by name, kind, or partial match. |
| `symbol_source` | Source for one function, class, or method without reading the whole file. |
| `find_references` | Textual references labeled by syntactic context: call, import, subclass, or attribute. |

The index is local-only. Per-repo SQLite databases live under `~/.cogni/<repo-hash>/index.db`.

## Commands

```sh
cogni install [--root PATH] [--local]      # configure Claude Code + index (run once per repo)
cogni serve [--root PATH] [--db PATH]      # MCP server over stdio
cogni index [PATH] [--stats] [--db PATH]   # build or refresh the index
cogni bench [--runs N] [--mcp-config PATH] # run the token-savings benchmark
cogni --version                            # print version
```

## How it works

1. **Parse**: tree-sitter (Python grammar, CGO-bound) extracts modules, classes, functions, methods, and their line ranges.
2. **Store**: symbols, files, and references are written to a per-repo SQLite database.
3. **Serve**: the MCP server exposes scoped queries over stdio so an agent can ask `repo_overview` instead of running `ls -R`, or `symbol_source` instead of reading a 600-line file.

Source layout:

```
cmd/cogni/         CLI entrypoint (serve, index, install, bench)
internal/parse/    tree-sitter Python parser
internal/index/    indexer + filesystem watcher
internal/store/    SQLite schema and queries
internal/mcp/      MCP tool definitions and handlers
internal/bench/    benchmark harness, scorers, task definitions
```

## Contributing

Contributions are welcome. The project uses standard Go tooling and a small set of conventions.

### Development setup

```sh
git clone https://github.com/islamborghini/cogni
cd cogni
go build -o cogni ./cmd/cogni
go test ./...
```

CGO is required (tree-sitter, SQLite). On macOS you need Xcode command line tools; on Linux a working `gcc` and `libc` headers.

### Before opening a PR

```sh
gofmt -w .
go vet ./...
go test ./...
go build -o ./cogni ./cmd/cogni
```

CI runs `fmt + vet + test` on Ubuntu and macOS, plus `golangci-lint`. Both must be green.

### Conventions

- **Branches**: `feat/`, `fix/`, `chore/`, `test/`, or `docs/` prefix.
- **PR titles**: same prefix as the branch, conventional-commit style.
- **Commits**: small, frequent, one-sentence conventional-commit messages.
- **Code**: comments only when the *why* isn't obvious; identifiers should explain the *what*.

### Where to start

- Open issues labeled `good first issue` or `help wanted`.
- New benchmark tasks in `internal/bench/tasks.yaml`. They must come from real merged PRs with checkable ground truth (see [BENCHMARK.md](BENCHMARK.md)).
- Bug reports with a minimal reproducer are very welcome.

## License

[Apache 2.0](LICENSE)
