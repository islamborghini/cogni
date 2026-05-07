package mcp

import (
	"context"
	"testing"

	mcppkg "github.com/mark3labs/mcp-go/mcp"
)

func TestSymbolSearchExact(t *testing.T) {
	_, _, c := fixture(t)
	out := callJSON(t, c, "symbol_search", map[string]any{"query": "hello"})
	matches, _ := out["matches"].([]any)
	if len(matches) == 0 {
		t.Fatalf("expected at least one match for 'hello', got: %v", out)
	}
	first := matches[0].(map[string]any)
	if first["name"] != "hello" {
		t.Errorf("first match name = %v, want hello", first["name"])
	}
	if first["path"] != "a.py" {
		t.Errorf("first match path = %v, want a.py", first["path"])
	}
}

func TestSymbolSearchTrigramFuzzy(t *testing.T) {
	_, _, c := fixture(t)
	// Trigram tokenizer should match "Greet" against "Greeter".
	out := callJSON(t, c, "symbol_search", map[string]any{"query": "Greet"})
	matches, _ := out["matches"].([]any)
	if len(matches) == 0 {
		t.Fatalf("expected fuzzy match for 'Greet', got: %v", out)
	}
	found := false
	for _, m := range matches {
		if m.(map[string]any)["name"] == "Greeter" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Greeter in matches, got: %v", matches)
	}
}

func TestSymbolSearchKindFilter(t *testing.T) {
	_, _, c := fixture(t)
	out := callJSON(t, c, "symbol_search", map[string]any{
		"query": "Greet", "kind": "class",
	})
	matches, _ := out["matches"].([]any)
	if len(matches) == 0 {
		t.Fatalf("expected class matches, got: %v", out)
	}
	for _, m := range matches {
		if k := m.(map[string]any)["kind"]; k != "class" {
			t.Errorf("kind filter leaked: got kind=%v", k)
		}
	}
}

func TestSymbolSearchMissingQueryErrors(t *testing.T) {
	_, _, c := fixture(t)
	req := mcppkg.CallToolRequest{}
	req.Params.Name = "symbol_search"
	req.Params.Arguments = map[string]any{}
	res, err := c.CallTool(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Errorf("expected error result for missing query")
	}
}
