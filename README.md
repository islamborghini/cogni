# Cogni

Open-core MCP (Model Context Protocol) server that drops into AI coding agents — Claude Code, Cursor, Windsurf, OpenCode — and reduces their token consumption by giving them structured recall over a repository.

> **Status:** pre-alpha. v0.1 in active development. See [What works today](#what-works-today).

## What it does

AI coding agents waste 40–60% of their token budget orienting in unfamiliar repos: re-reading files, grepping for symbols, accumulating stale context. Cogni indexes the repo into a code knowledge graph and exposes it through MCP tools so the agent can drill from overview → package → symbol → source instead of brute-forcing.

v0.1 ships **structural retrieval** (Layer 1) for Python repositories. Pruning (Layer 2) and autonomous checkpoints (Layer 3) are planned for v0.2+.

## Install

Requires Go 1.26+ (a single static binary via Homebrew and GitHub Releases is planned for v0.1 release).

```sh
go install github.com/islamborghini/cogni/cmd/cogni@latest
```

Or from source:

```sh
git clone https://github.com/islamborghini/cogni
cd cogni
go build -o cogni ./cmd/cogni
```

## Quickstart

Index a Python repo and inspect it:

```sh
cd path/to/your/python/repo
cogni index . --stats
```

Expected output:

```
indexed 142 files, 1873 symbols in 412ms
db: ~/.cogni/<repo-hash>/index.db (3.2 MB)
```

Run the MCP server:

```sh
cogni serve --root path/to/your/python/repo
```

`cogni serve` speaks JSON-RPC over stdio. It auto-indexes on startup and watches for file changes.

## Wire into Claude Code

Add cogni to your MCP server config (`~/.claude.json` or per-project `.mcp.json`):

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

Restart Claude Code; the five Cogni tools will appear in the tool list.

## What works today

v0.1 is mid-development. Currently real:

- `repo_overview` — package tree with file counts and top exported symbols
- `file_outline` — symbols defined in a single file with line ranges and signatures

Currently stubbed (return `{"stub": true}`; real implementations land in week 2):

- `symbol_search`, `symbol_source`, `find_references`

Indexing covers Python only.

## Commands

```sh
cogni serve [--root PATH] [--db PATH]   # MCP server over stdio
cogni index PATH [--stats]              # explicit index build
```

`cogni bench` (token-savings harness) ships in week 4.

## License

[Apache 2.0](LICENSE)
