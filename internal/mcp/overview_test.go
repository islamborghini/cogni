package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/islamborghini/cogni/internal/index"
	"github.com/islamborghini/cogni/internal/store"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcppkg "github.com/mark3labs/mcp-go/mcp"
)

// fixture builds a tiny repo and returns root, store, and started in-process client.
func fixture(t *testing.T) (string, *store.Store, *mcpclient.Client) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"a.py": "def hello():\n    return 1\n\ndef _private():\n    return 2\n",
		"pkg/b.py": "class Greeter:\n    def greet(self):\n        return 'hi'\n",
		"pkg/sub/c.py": "def worker():\n    return 3\n",
	}
	for rel, body := range files {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	s, err := store.Open(filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if _, err := index.Build(root, s); err != nil {
		t.Fatal(err)
	}

	srv := New(s)
	c, err := mcpclient.NewInProcessClient(srv.MCP())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	if err := c.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	init := mcppkg.InitializeRequest{}
	init.Params.ProtocolVersion = mcppkg.LATEST_PROTOCOL_VERSION
	init.Params.ClientInfo = mcppkg.Implementation{Name: "test", Version: "0"}
	if _, err := c.Initialize(context.Background(), init); err != nil {
		t.Fatal(err)
	}
	return root, s, c
}

func callJSON(t *testing.T, c *mcpclient.Client, name string, args map[string]any) map[string]any {
	t.Helper()
	req := mcppkg.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	res, err := c.CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("CallTool %s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s returned error: %+v", name, res.Content)
	}
	text := res.Content[0].(mcppkg.TextContent).Text
	var out map[string]any
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("invalid json from %s: %v\n%s", name, err, text)
	}
	return out
}

func TestRepoOverviewReturnsTree(t *testing.T) {
	_, _, c := fixture(t)
	out := callJSON(t, c, "repo_overview", map[string]any{"max_depth": 3})

	pkgs, _ := out["packages"].([]any)
	if len(pkgs) == 0 {
		t.Fatalf("expected packages, got: %v", out)
	}
	// Expect at least these prefixes: ".", "pkg", "pkg/sub"
	want := map[string]bool{".": false, "pkg": false, "pkg/sub": false}
	for _, p := range pkgs {
		m := p.(map[string]any)
		if _, ok := want[m["path"].(string)]; ok {
			want[m["path"].(string)] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing package %q in overview", k)
		}
	}

	// "pkg" should surface Greeter (a class), not _private.
	for _, p := range pkgs {
		m := p.(map[string]any)
		if m["path"] != "pkg" {
			continue
		}
		syms, _ := m["top_symbols"].([]any)
		foundGreeter := false
		for _, sm := range syms {
			s := sm.(map[string]any)
			if s["name"] == "Greeter" {
				foundGreeter = true
			}
			if s["name"] == "_private" {
				t.Errorf("underscore symbol leaked into top_symbols")
			}
		}
		if !foundGreeter {
			t.Errorf("expected Greeter under pkg, got: %v", syms)
		}
	}
}

func TestFileOutlineReturnsSymbols(t *testing.T) {
	_, _, c := fixture(t)
	out := callJSON(t, c, "file_outline", map[string]any{"path": "pkg/b.py"})
	if out["path"] != "pkg/b.py" {
		t.Errorf("path mismatch: %v", out["path"])
	}
	syms, _ := out["symbols"].([]any)
	if len(syms) < 2 {
		t.Fatalf("expected ≥2 symbols (class + method), got %d: %v", len(syms), syms)
	}
	first := syms[0].(map[string]any)
	if first["name"] != "Greeter" {
		t.Errorf("first symbol should be Greeter, got %v", first["name"])
	}
}

func TestFileOutlineMissingPathErrors(t *testing.T) {
	_, _, c := fixture(t)
	req := mcppkg.CallToolRequest{}
	req.Params.Name = "file_outline"
	req.Params.Arguments = map[string]any{}
	res, err := c.CallTool(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Errorf("expected error result for missing path")
	}
}
