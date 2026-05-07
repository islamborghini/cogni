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
			"Get a high-level map of this repository before reading any files. "+
				"Returns the package/module tree with file counts, a one-line summary per package, "+
				"and the top exported symbols. Call this FIRST when you need to orient in an "+
				"unfamiliar repo or answer 'where would X live'. Avoid reading multiple files "+
				"just to figure out the layout — this tool is cheaper.",
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
			"List the symbols (classes, functions, methods, top-level constants) defined in a "+
				"single file with line ranges and signatures, WITHOUT returning the source code. "+
				"Use this when you need to know what's in a file before deciding whether to read it. "+
				"Prefer this over reading the whole file when the question is 'what does this file "+
				"expose' or 'is X defined here'.",
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Repo-relative path."),
		),
	)
}

func symbolSearchTool() mcp.Tool {
	return mcp.NewTool("symbol_search",
		mcp.WithDescription(
			"Find symbol definitions by name across the whole repo. Returns file path, line range, "+
				"kind (function/class/method), and signature for each match. Use this instead of grep "+
				"when you're looking for WHERE something is defined. Supports exact, prefix, and fuzzy "+
				"matches. Example: query='register_blueprint' → returns every definition with that name.",
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
			"Return the source code of a single symbol (function, class, or method) by name or "+
				"qualified name. Returns just the symbol's lines, not the whole file. Use this after "+
				"symbol_search when you need to actually read an implementation. Pass "+
				"qualified='flask.app.Flask.register_blueprint' to disambiguate when multiple symbols "+
				"share a name.",
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
			"Find places that reference a symbol by name (call sites, attribute access, imports, "+
				"subclassing). Returns file:line for each occurrence with the surrounding line as "+
				"context. Use this instead of grep when answering 'where is X used' or 'what calls Y'. "+
				"Note: v0.1 uses textual matching, so common names may include false positives — the "+
				"tool labels each result with its syntactic context (call/attribute/import/subclass).",
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
		Docstring string `json:"docstring,omitempty"`
	}
	out := make([]outlineSym, 0, len(rows))
	for _, r := range rows {
		out = append(out, outlineSym{
			Name: r.Name, Qualified: r.Qualified, Kind: r.Kind,
			StartLine: r.StartLine, EndLine: r.EndLine,
			Signature: r.Signature, Docstring: r.Docstring,
		})
	}
	return jsonResult(map[string]any{"path": path, "symbols": out})
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
	return jsonResult(map[string]any{
		"stub":       true,
		"name":       req.GetString("name", ""),
		"references": []any{},
	})
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
