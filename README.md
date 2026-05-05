# Cogni

Open-core MCP (Model Context Protocol) server that drops into AI coding agents — Claude Code, Cursor, Windsurf, OpenCode — and reduces their token consumption by giving them structured recall over a repository.

> **Status:** pre-alpha. v0.1 in active development.

## What it does

AI coding agents waste 40–60% of their token budget orienting in unfamiliar repos: re-reading files, grepping for symbols, accumulating stale context. Cogni indexes the repo into a code knowledge graph and exposes it through MCP tools so the agent can drill from overview → package → symbol → source instead of brute-forcing.

v0.1 ships **structural retrieval** (Layer 1) for Python repositories. Pruning (Layer 2) and autonomous checkpoints (Layer 3) are planned for v0.2+.

## Install

_Coming soon._ A single static binary will be available via Homebrew and GitHub Releases.

## Usage

```sh
cogni serve         # run the MCP server over stdio
cogni index .       # index the current repository
cogni bench         # run the token-savings benchmark
```

## License

[Apache 2.0](LICENSE)
