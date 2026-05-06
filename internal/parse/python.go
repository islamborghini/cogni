package parse

import (
	"errors"

	ts "github.com/tree-sitter/go-tree-sitter"
	tspython "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// pythonLang is loaded once. tree-sitter Languages are immutable and safe for
// concurrent reuse, unlike Parser instances.
var pythonLang = ts.NewLanguage(tspython.Language())

// ParsePython returns every named definition (function, class, method) in src.
// Qualified names are dotted paths within the file: a top-level function gets
// Qualified == Name; a method on Greeter gets Qualified == "Greeter.greet".
//
// ParsePythonTopLevel is preserved for callers that only want module-level
// items; it is now a thin filter over ParsePython.
func ParsePython(src []byte) ([]Symbol, error) {
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

	var out []Symbol
	walkScope(tree.RootNode(), src, "", false, &out)
	return out, nil
}

// ParsePythonTopLevel returns only the module-level functions and classes.
func ParsePythonTopLevel(src []byte) ([]Symbol, error) {
	all, err := ParsePython(src)
	if err != nil {
		return nil, err
	}
	out := all[:0]
	for _, s := range all {
		if s.Kind == KindFunction || s.Kind == KindClass {
			if s.Qualified == s.Name { // no dotted prefix → top-level
				out = append(out, s)
			}
		}
	}
	return out, nil
}

// walkScope visits a node's named children and emits symbols. parent is the
// qualified-name prefix ("" at module level, "Greeter" inside a class body).
// inClass flips KindFunction → KindMethod for definitions found one level
// inside a class.
func walkScope(scope *ts.Node, src []byte, parent string, inClass bool, out *[]Symbol) {
	for i := uint(0); i < scope.NamedChildCount(); i++ {
		c := scope.NamedChild(i)
		switch c.Kind() {
		case "function_definition", "decorated_definition", "class_definition":
			sym, body, ok := extractDef(c, src, parent)
			if !ok {
				continue
			}
			if inClass && sym.Kind == KindFunction {
				sym.Kind = KindMethod
			}
			*out = append(*out, sym)
			if sym.Kind == KindClass && body != nil {
				walkScope(body, src, sym.Qualified, true, out)
			}
		}
	}
}

// extractDef pulls a Symbol out of a function_definition, class_definition, or
// decorated_definition node and also returns the body block (for classes, so
// the caller can recurse). ok=false for unsupported node kinds or malformed
// definitions missing a name.
func extractDef(n *ts.Node, src []byte, parent string) (Symbol, *ts.Node, bool) {
	startLine := int(n.StartPosition().Row) + 1
	endLine := int(n.EndPosition().Row) + 1

	def := n
	if n.Kind() == "decorated_definition" {
		def = n.ChildByFieldName("definition")
		if def == nil {
			return Symbol{}, nil, false
		}
	}

	var kind SymbolKind
	switch def.Kind() {
	case "function_definition":
		kind = KindFunction
	case "class_definition":
		kind = KindClass
	default:
		return Symbol{}, nil, false
	}

	nameNode := def.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}, nil, false
	}
	name := nameNode.Utf8Text(src)
	qualified := name
	if parent != "" {
		qualified = parent + "." + name
	}

	sym := Symbol{
		Name:      name,
		Qualified: qualified,
		Kind:      kind,
		StartLine: startLine,
		EndLine:   endLine,
	}

	body := def.ChildByFieldName("body")
	if kind == KindClass {
		return sym, body, true
	}
	return sym, nil, true
}
