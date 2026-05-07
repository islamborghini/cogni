package mcp

import (
	"context"
	"strings"
	"testing"

	mcppkg "github.com/mark3labs/mcp-go/mcp"
)

func TestSymbolSourceByName(t *testing.T) {
	_, _, c := fixture(t)
	out := callJSON(t, c, "symbol_source", map[string]any{"name": "hello"})
	if out["path"] != "a.py" {
		t.Errorf("path = %v, want a.py", out["path"])
	}
	src, _ := out["source"].(string)
	if !strings.Contains(src, "def hello") {
		t.Errorf("source missing 'def hello':\n%s", src)
	}
	if out["start_line"].(float64) != 1 {
		t.Errorf("start_line = %v, want 1", out["start_line"])
	}
}

func TestSymbolSourceContextLines(t *testing.T) {
	_, _, c := fixture(t)
	// hello is at line 1; context_lines=1 should still start at 1 (clamped).
	out := callJSON(t, c, "symbol_source", map[string]any{
		"name": "hello", "context_lines": 1,
	})
	if out["start_line"].(float64) != 1 {
		t.Errorf("start_line clamp failed: %v", out["start_line"])
	}
	src, _ := out["source"].(string)
	// hello body is 2 lines; +1 context after should pull in the blank/next line.
	if strings.Count(src, "\n") < 2 {
		t.Errorf("expected ≥2 lines with context, got:\n%s", src)
	}
}

func TestSymbolSourceNotFound(t *testing.T) {
	_, _, c := fixture(t)
	out := callJSON(t, c, "symbol_source", map[string]any{"name": "no_such_symbol"})
	matches, _ := out["matches"].([]any)
	if len(matches) != 0 {
		t.Errorf("expected empty matches, got: %v", matches)
	}
	if _, ok := out["source"]; ok {
		t.Errorf("unexpected source field on miss")
	}
}

func TestSymbolSourceRequiresNameOrQualified(t *testing.T) {
	_, _, c := fixture(t)
	req := mcppkg.CallToolRequest{}
	req.Params.Name = "symbol_source"
	req.Params.Arguments = map[string]any{}
	res, err := c.CallTool(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Errorf("expected error result for missing name/qualified")
	}
}
