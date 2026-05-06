package parse

import (
	"errors"

	ts "github.com/tree-sitter/go-tree-sitter"
	tspython "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// pythonLang is loaded once. tree-sitter Languages are immutable and safe for
// concurrent reuse, unlike Parser instances.
var pythonLang = ts.NewLanguage(tspython.Language())

// ParsePythonTopLevel returns the top-level function and class definitions in
// src. Nested definitions (methods inside a class, inner functions) are not
// returned here — Tue AM extends the parser to walk into class bodies and
// emit KindMethod symbols.
func ParsePythonTopLevel(src []byte) ([]Symbol, error) {
	parser := ts.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(pythonLang); err != nil {
		return nil, err
	}
	tree := parser.Parse(src, nil)
	if tree == nil {
		return nil, errors.New("parse: tree-sitter returned nil tree")
	}
	defer tree.Close()

	root := tree.RootNode()
	var out []Symbol
	for i := uint(0); i < root.NamedChildCount(); i++ {
		c := root.NamedChild(i)
		if sym, ok := topLevelSymbol(c, src); ok {
			out = append(out, sym)
		}
	}
	return out, nil
}

// topLevelSymbol pulls a Symbol out of a top-level tree-sitter node, or
// returns ok=false if the node isn't a definition we surface (assignments,
// imports, docstrings, etc.).
//
// A `decorated_definition` wraps a `function_definition` or `class_definition`
// when the user prefixed it with one or more `@decorator` lines; we unwrap to
// the inner definition for kind/name but keep the outer line range so the
// decorators stay associated with the symbol.
func topLevelSymbol(n *ts.Node, src []byte) (Symbol, bool) {
	startLine := int(n.StartPosition().Row) + 1
	endLine := int(n.EndPosition().Row) + 1

	def := n
	if n.Kind() == "decorated_definition" {
		def = n.ChildByFieldName("definition")
		if def == nil {
			return Symbol{}, false
		}
	}

	var kind SymbolKind
	switch def.Kind() {
	case "function_definition":
		kind = KindFunction
	case "class_definition":
		kind = KindClass
	default:
		return Symbol{}, false
	}

	nameNode := def.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}, false
	}
	name := nameNode.Utf8Text(src)
	return Symbol{
		Name:      name,
		Qualified: name,
		Kind:      kind,
		StartLine: startLine,
		EndLine:   endLine,
	}, true
}
