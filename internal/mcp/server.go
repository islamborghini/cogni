// Package mcp wires the Cogni MCP server. v0.1 ships five tools that let
// coding agents recall repo structure without grep/cat: repo_overview,
// file_outline, symbol_search, symbol_source, find_references.
//
// Handlers are stubs in Thu AM — descriptions are the load-bearing thing here
// (they steer agent tool selection). Real handlers wire to internal/store
// starting Thu PM.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/islamborghini/cogni/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Version is reported to MCP clients during initialize.
const Version = "0.1.0-dev"

// Server holds the MCP server bound to a Cogni store and the repo root
// (used by tools that read source from disk, e.g. symbol_source).
type Server struct {
	mcp   *server.MCPServer
	store *store.Store
	root  string
}

// New constructs an MCP server with all five Cogni tools registered. Root
// must be the absolute path to the repository being served.
func New(s *store.Store, root string) *Server {
	m := server.NewMCPServer(
		"cogni",
		Version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	srv := &Server{mcp: m, store: s, root: root}
	srv.registerTools()
	return srv
}

// ServeStdio runs the JSON-RPC loop over stdin/stdout. It blocks until the
// client disconnects.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcp)
}

// MCP exposes the underlying server for tests.
func (s *Server) MCP() *server.MCPServer { return s.mcp }

func (s *Server) registerTools() {
	s.mcp.AddTool(repoOverviewTool(), s.handleRepoOverview)
	s.mcp.AddTool(fileOutlineTool(), s.handleFileOutline)
	s.mcp.AddTool(symbolSearchTool(), s.handleSymbolSearch)
	s.mcp.AddTool(symbolSourceTool(), s.handleSymbolSource)
	s.mcp.AddTool(findReferencesTool(), s.handleFindReferences)
}

func repoOverviewTool() mcp.Tool {
	return mcp.NewTool("repo_overview",
		mcp.WithDescription(
			"Returns the repository's package/module tree with file counts and top exported symbols. "+
				"Useful for initial orientation when you do not know where a feature lives. "+
				"Skip it if you already know which package to look in — use symbol_search directly instead.",
		),
		mcp.WithNumber("max_depth",
			mcp.Description("Package tree depth."),
			mcp.DefaultNumber(3),
		),
	)
}

func fileOutlineTool() mcp.Tool {
	return mcp.NewTool("file_outline",
		mcp.WithDescription(
			"List the symbols (classes, functions, methods, constants) in a file with line ranges and signatures. "+
				"Best for small-to-medium files where you need a complete inventory. "+
				"For large files or when you are looking for a specific symbol, prefer symbol_search + symbol_source — "+
				"they return only the relevant item rather than the full symbol list. "+
				"After this, call symbol_source for any specific symbol you want to inspect — never Read the whole file.",
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Repo-relative path."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum symbols to return. Default 60; raise to 200 if you need the full inventory."),
			mcp.DefaultNumber(60),
			mcp.Max(200),
		),
	)
}

func symbolSearchTool() mcp.Tool {
	return mcp.NewTool("symbol_search",
		mcp.WithDescription(
			"USE THIS INSTEAD OF Grep for finding WHERE a symbol is DEFINED. "+
				"Returns file path, line range, kind (function/class/method), and signature for each definition. "+
				"Indexes the repo so it's exact and fast — Grep returns every textual mention including "+
				"comments and call sites, which is the wrong answer when you want the definition. "+
				"Trigger phrases: 'where is X defined', 'find the X function', 'show me the Y class'. "+
				"Supports fuzzy/partial names. Example: query='register_blueprint' returns every definition.",
		),
		mcp.WithString("query", mcp.Required()),
		mcp.WithString("kind",
			mcp.Description("function|class|method|variable|constant|any"),
			mcp.DefaultString("any"),
		),
		mcp.WithNumber("limit", mcp.DefaultNumber(20), mcp.Max(100)),
	)
}

func symbolSourceTool() mcp.Tool {
	return mcp.NewTool("symbol_source",
		mcp.WithDescription(
			"USE THIS INSTEAD OF Read when you want one specific function, class, or method's body. "+
				"Returns just that symbol's lines — typically 5-50 lines vs the whole file's 200-2000. "+
				"The natural follow-up after symbol_search: search returns where, this returns the code. "+
				"Trigger phrases: 'show me the implementation of X', 'how does function Y work', "+
				"'read the X method'. "+
				"Pass qualified='module.Class.method' to disambiguate when multiple symbols share a name.",
		),
		mcp.WithString("name",
			mcp.Description("Bare name; ambiguous matches return a disambiguation list."),
		),
		mcp.WithString("qualified",
			mcp.Description("Fully-qualified name; preferred when known."),
		),
		mcp.WithNumber("context_lines", mcp.DefaultNumber(0), mcp.Max(20)),
	)
}

func findReferencesTool() mcp.Tool {
	return mcp.NewTool("find_references",
		mcp.WithDescription(
			"USE THIS INSTEAD OF Grep for finding USES of a symbol (call sites, imports, subclassing). "+
				"Returns file:line for each occurrence labeled by syntactic context "+
				"(call/import/subclass) — Grep returns raw text matches that you'd have to filter manually. "+
				"Trigger phrases: 'where is X used', 'what calls Y', 'who imports Z', 'find usages of X', "+
				"'rename X across the repo' (call this first, then edit each site). "+
				"v0.1 caveat: textual matching, so common names may include false positives — the kind "+
				"label tells you which results are real call sites vs other contexts.",
		),
		mcp.WithString("name", mcp.Required()),
		mcp.WithArray("kinds",
			mcp.Description("Filter to specific reference kinds."),
			mcp.Items(map[string]any{
				"type": "string",
				"enum": []string{"call", "attribute", "import", "subclass"},
			}),
		),
		mcp.WithNumber("limit", mcp.DefaultNumber(50), mcp.Max(200)),
	)
}

// --- stub handlers -------------------------------------------------------
// Each returns canned JSON shaped like the real response so we can validate
// agent tool-selection behavior before implementing the queries.

func (s *Server) handleRepoOverview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	maxDepth := int(req.GetFloat("max_depth", 3))
	pkgs, err := BuildOverview(s.store, maxDepth, 5)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return jsonResult(map[string]any{"packages": pkgs})
}

func (s *Server) handleFileOutline(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := req.GetString("path", "")
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}
	limit := int(req.GetFloat("limit", 60))
	if limit <= 0 {
		limit = 60
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.store.SymbolsByFile(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	type outlineSym struct {
		Name      string `json:"name"`
		Qualified string `json:"qualified"`
		Kind      string `json:"kind"`
		StartLine int    `json:"start_line"`
		EndLine   int    `json:"end_line"`
		Signature string `json:"signature,omitempty"`
	}
	total := len(rows)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]outlineSym, 0, len(rows))
	for _, r := range rows {
		sig := r.Signature
		if len(sig) > 100 {
			sig = sig[:97] + "..."
		}
		out = append(out, outlineSym{
			Name: r.Name, Qualified: r.Qualified, Kind: r.Kind,
			StartLine: r.StartLine, EndLine: r.EndLine,
			Signature: sig,
		})
	}
	result := map[string]any{"path": path, "symbols": out, "total_symbols": total}
	if total > limit {
		result["truncated"] = true
		result["hint"] = "use symbol_search to find specific items in this large file"
	}
	return jsonResult(result)
}

func (s *Server) handleSymbolSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q := strings.TrimSpace(req.GetString("query", ""))
	if q == "" {
		return mcp.NewToolResultError("query is required"), nil
	}
	kind := req.GetString("kind", "any")
	limit := int(req.GetFloat("limit", 20))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Wrap as an FTS5 phrase so tokenizer specials (`_`, `:`, `*`, parens,
	// `AND`/`OR`/`NOT`) in user input can't break the MATCH grammar. Inner
	// quotes are escaped by doubling.
	ftsQ := `"` + strings.ReplaceAll(q, `"`, `""`) + `"`

	rows, err := s.store.SymbolsFTS(ftsQ, kind, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	type match struct {
		Name      string `json:"name"`
		Qualified string `json:"qualified"`
		Kind      string `json:"kind"`
		Path      string `json:"path"`
		StartLine int    `json:"start_line"`
		EndLine   int    `json:"end_line"`
		Signature string `json:"signature,omitempty"`
	}
	out := make([]match, 0, len(rows))
	for _, r := range rows {
		out = append(out, match{
			Name: r.Name, Qualified: r.Qualified, Kind: r.Kind,
			Path: r.FilePath, StartLine: r.StartLine, EndLine: r.EndLine,
			Signature: r.Signature,
		})
	}
	return jsonResult(map[string]any{"query": q, "kind": kind, "matches": out})
}

func (s *Server) handleSymbolSource(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := strings.TrimSpace(req.GetString("name", ""))
	qualified := strings.TrimSpace(req.GetString("qualified", ""))
	if name == "" && qualified == "" {
		return mcp.NewToolResultError("name or qualified is required"), nil
	}
	contextLines := int(req.GetFloat("context_lines", 0))
	if contextLines < 0 {
		contextLines = 0
	}
	if contextLines > 20 {
		contextLines = 20
	}

	var rows []store.SymbolRow
	var err error
	if qualified != "" {
		rows, err = s.store.SymbolsByQualified(qualified, 10)
	} else {
		rows, err = s.store.SymbolsByName(name, 10)
	}
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if len(rows) == 0 {
		return jsonResult(map[string]any{
			"name": name, "qualified": qualified, "matches": []any{},
		})
	}
	if len(rows) > 1 {
		// Ambiguous bare-name lookup: return disambiguation list, no source.
		type cand struct {
			Name      string `json:"name"`
			Qualified string `json:"qualified"`
			Kind      string `json:"kind"`
			Path      string `json:"path"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		}
		out := make([]cand, 0, len(rows))
		for _, r := range rows {
			out = append(out, cand{r.Name, r.Qualified, r.Kind, r.FilePath, r.StartLine, r.EndLine})
		}
		return jsonResult(map[string]any{
			"name":      name,
			"qualified": qualified,
			"ambiguous": true,
			"matches":   out,
		})
	}

	r := rows[0]
	src, startOut, endOut, err := readLineRange(s.root, r.FilePath, r.StartLine, r.EndLine, contextLines)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return jsonResult(map[string]any{
		"name":       r.Name,
		"qualified":  r.Qualified,
		"kind":       r.Kind,
		"path":       r.FilePath,
		"start_line": startOut,
		"end_line":   endOut,
		"signature":  r.Signature,
		"source":     src,
	})
}

// readLineRange returns the slice of lines [start-context, end+context] from
// repo-relative path, clamped to file bounds. The returned start/end reflect
// the actual span emitted (1-indexed).
func readLineRange(root, relPath string, start, end, context int) (string, int, int, error) {
	full := filepath.Join(root, relPath)
	f, err := os.Open(full)
	if err != nil {
		return "", 0, 0, err
	}
	defer f.Close()
	from := start - context
	if from < 1 {
		from = 1
	}
	to := end + context
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var b strings.Builder
	line := 0
	for scanner.Scan() {
		line++
		if line < from {
			continue
		}
		if line > to {
			break
		}
		b.WriteString(scanner.Text())
		b.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", 0, 0, err
	}
	if line < to {
		to = line
	}
	return b.String(), from, to, nil
}

func (s *Server) handleFindReferences(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := strings.TrimSpace(req.GetString("name", ""))
	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}
	limit := int(req.GetFloat("limit", 50))
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	kinds := req.GetStringSlice("kinds", nil)

	rows, err := s.store.RefsByName(name, kinds, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	type refOut struct {
		Path        string `json:"path"`
		Line        int    `json:"line"`
		Col         int    `json:"col"`
		ContextKind string `json:"context_kind"`
	}
	out := make([]refOut, 0, len(rows))
	for _, r := range rows {
		out = append(out, refOut{r.FilePath, r.Line, r.Col, r.ContextKind})
	}
	return jsonResult(map[string]any{
		"name":       name,
		"references": out,
	})
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
