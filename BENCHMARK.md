# Cogni v0.1 Benchmark Methodology

This document defines how Cogni's token-savings benchmark is run, scored, and
reproduced. The launch claim — "Cogni reduces agent token consumption by ≥30%
on real Python tasks" — is only as defensible as this methodology, so it is
versioned alongside the code.

The actual task definitions live in [`internal/bench/tasks.yaml`](internal/bench/tasks.yaml).

## What the benchmark measures

For each task, an agent attempts to complete the task **twice** under
otherwise-identical conditions:

- **Baseline**: agent has its default tools only (Read, Grep, Bash, …).
- **With Cogni**: same agent + the five Cogni MCP tools (`repo_overview`,
  `file_outline`, `symbol_search`, `symbol_source`, `find_references`).

We measure:

- **Total input + output tokens** (read from the API's response metadata, not
  estimated client-side).
- **Pass/fail** against a per-task rubric (see [Scoring](#scoring)).
- **Wall-clock duration** (informational only; not part of the launch claim).

Headline number: **mean token reduction across passing runs**, reported with
n=5 per condition per task.

## Fixed inputs (must not vary across runs)

| Input | Value |
|---|---|
| Target repo | `httpx` at commit `<PINNED_SHA>` (TBD; pinned before week-4 measurement) |
| Agent | Claude Code SDK in headless mode |
| Model | `claude-sonnet-4-6` (or whatever the SDK defaults to at measurement time — recorded per run) |
| System prompt | Verbatim, checked into `internal/bench/system_prompt.md` |
| Tool registration | Either default (baseline) or default + Cogni MCP server (with-Cogni) |
| Task list | `internal/bench/tasks.yaml`, locked at the commit that produces `bench-report.md` |
| Runs per condition | n=5 |
| Indexing | `cogni index .` runs once before each with-Cogni session; cost is reported separately |

## Task selection criteria

Tasks must:

1. **Have a ground-truth answer**. We draw from real merged PRs and historical
   issues so success is checkable, not subjective.
2. **Span agent failure modes Cogni is designed to fix**: orienting in an
   unfamiliar package, locating where a symbol is defined, finding callers,
   reading a single function without pulling whole files.
3. **Cover four task families** (≥2 per family for n=8 total):
   - **explain** — "How is X wired? What does Y do?"
   - **bug-fix** — reproduce a known fixed bug from git history.
   - **add-feature** — implement a small change matching a known PR.
   - **refactor** — rename/restructure across multiple call sites.

## Scoring

Each run is graded against the task's `success_criteria` field, which is a
list of one or more checkable assertions:

- `output_contains: "..."` — string must appear in the agent's final answer.
- `file_modified: <path>` — file must be modified in the agent's workspace.
- `tests_pass: <pytest_node>` — when the task references existing repo tests.
- `function_added: <module>.<name>` — for add-feature tasks.

A run is a **pass** iff every criterion is met. Token counts from failed runs
are reported but excluded from the headline mean.

## Reproduction

```sh
# 1. Clone, pin, install
git clone https://github.com/encode/httpx /tmp/httpx
git -C /tmp/httpx checkout <PINNED_SHA>

# 2. Build cogni at the same commit that produced bench-report.md
go install github.com/islamborghini/cogni/cmd/cogni@<COGNI_SHA>

# 3. Run
cogni bench \
  --tasks internal/bench/tasks.yaml \
  --target /tmp/httpx \
  --runs 5 \
  --output bench-report.md
```

`cogni bench` writes a markdown report with: per-task mean tokens (baseline
vs with-cogni), pass-rate, and the full per-run breakdown for audit.

## Sanity check

In addition to the SDK-based runs, n=2 tasks are re-run via a thin direct
Anthropic API loop (`internal/bench/direct.go`) to confirm the direction of
the effect isn't an artifact of the Claude Code SDK's tool-selection
heuristics. Direct-API numbers are reported alongside but do not replace the
headline.

## Known threats to validity

- **SDK behavior drift**: a future SDK version may change tool-selection
  defaults. We record SDK version per run and re-pin if numbers shift
  materially.
- **Indexing cost**: amortizing index time across n=5 runs is favorable to
  Cogni; we report indexing cost separately so reviewers can include it.
- **Task overfit**: tasks were chosen knowing what Cogni does. Mitigation:
  tasks are drawn from real historical work, not synthesized for the
  benchmark. The `tasks.yaml` commit history shows exactly when each was
  added.
- **Single repo**: v0.1 measures only on httpx. v0.2 will add Flask and
  pandas to test transfer.

## v0.2 follow-ups (not in v0.1)

- Multi-repo runs.
- Pass-rate as a primary metric alongside tokens.
- Latency breakdown (tool-call vs model time).
- Per-tool ablation (which Cogni tool contributes the most savings).
