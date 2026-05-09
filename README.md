# Cogni

Open-core MCP (Model Context Protocol) server that drops into AI coding agents - Claude Code, Cursor, Windsurf, OpenCode - and reduces token consumption by giving them structured recall over a repository.

> **Status:** v0.1 pre-release. Python structural retrieval is implemented; pruning and autonomous checkpoints are planned for v0.2+.

## What it does

AI coding agents waste a large share of their token budget orienting in unfamiliar repos: re-reading files, grepping for symbols, and accumulating stale context. Cogni indexes a Python repository into a local SQLite code graph and exposes scoped MCP tools so agents can move from overview -> file -> symbol -> source instead of brute-forcing context.

v0.1 ships **Layer 1: structural retrieval** for Python repositories.

## Install

Requires Go 1.26+.

```sh
go install github.com/islamborghini/cogni/cmd/cogni@latest
```

Or build from source:

```sh
git clone https://github.com/islamborghini/cogni
cd cogni
go build -o cogni ./cmd/cogni
```

Prebuilt binaries and Homebrew installation are planned for the v0.1 release.

## Quickstart

Index a Python repo:

```sh
cd path/to/your/python/repo
cogni index . --stats
```

Run the MCP server:

```sh
cogni serve --root /absolute/path/to/your/python/repo
```

`cogni serve` speaks JSON-RPC over stdio. It auto-indexes on startup and watches for file changes.

## Using with Claude Code

Claude Code can connect to Cogni through MCP, but the MCP registration alone is not enough to reliably change tool selection. Add a `CLAUDE.md` file in your repo root so Claude Code prefers Cogni's scoped tools over broad Read/Grep/Glob exploration.

1. Add Cogni to your Claude Code MCP config (`~/.claude.json` or per-project `.mcp.json`):

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

2. Add this `CLAUDE.md` to the root of the repository you want Claude Code to work in:

```markdown
# Repository tooling

This repository has the **cogni** MCP server registered with these tools:

- `mcp__cogni__repo_overview` — high-level package map. Call FIRST for any architecture / "how does X work" question.
- `mcp__cogni__file_outline` — list symbols in a file without returning source. Use INSTEAD OF Read when you only need to know what a file exposes.
- `mcp__cogni__symbol_search` — find where a symbol is defined. Use INSTEAD OF Grep for definition lookups.
- `mcp__cogni__symbol_source` — return one symbol's body. Use INSTEAD OF Read after symbol_search.
- `mcp__cogni__find_references` — find usages of a symbol. Use INSTEAD OF Grep when answering "where is X used".

For code exploration tasks, **prefer the cogni tools** over Glob / Grep / Read. They return structured, scoped results and use 5-10x fewer tokens for the same answer.
```

3. Restart Claude Code and work in that repository as usual.

## Tools

Cogni exposes five MCP tools:

- `repo_overview` - package/module tree with file counts and top symbols.
- `file_outline` - symbols in one file with line ranges, signatures, and docstrings.
- `symbol_search` - indexed definition search by name, kind, or partial match.
- `symbol_source` - source for one function, class, or method without reading the whole file.
- `find_references` - textual v0.1 references labeled by syntactic context: call, import, subclass, or attribute.

The v0.1 index is local-only. By default, per-repo SQLite databases live under `~/.cogni/<repo-hash>/index.db`.

## Commands

```sh
cogni serve [--root PATH] [--db PATH]      # MCP server over stdio
cogni index [PATH] [--stats] [--db PATH]   # build or refresh the index
cogni bench [--runs N] [--mcp-config PATH] # run the token-savings benchmark
```

## Benchmark

The benchmark harness runs the same task set with and without Cogni tools, then reports token usage and pass/fail results. See [BENCHMARK.md](BENCHMARK.md) for methodology.

Current local launch measurement: **24.3% token reduction** on the `explain-transports` httpx task across n=5 runs, with 5/5 pass-rate in both conditions. The broader v0.1 benchmark task set is still being finalized.

## License

[Apache 2.0](LICENSE)
