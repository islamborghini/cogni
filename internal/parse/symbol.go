// Package parse extracts symbols and references from source files.
//
// v0.1 supports Python only, via the pure-Go tree-sitter runtime in
// github.com/odvcencio/gotreesitter (grammars are embedded, so there is
// nothing to vendor).
package parse

// SymbolKind classifies a parsed symbol. Values are stable strings because
// they are persisted to the on-disk index and surfaced through MCP tools.
type SymbolKind string

const (
	KindFunction SymbolKind = "function"
	KindClass    SymbolKind = "class"
	KindMethod   SymbolKind = "method"
	KindVariable SymbolKind = "variable"
	KindConstant SymbolKind = "constant"
)

// Symbol is a single named definition extracted from a source file.
// Line numbers are 1-based and inclusive, matching what users see in editors.
type Symbol struct {
	Name      string
	Kind      SymbolKind
	StartLine int
	EndLine   int
}
