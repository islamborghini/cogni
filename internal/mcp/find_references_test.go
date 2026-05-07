package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/islamborghini/cogni/internal/index"
	"github.com/islamborghini/cogni/internal/store"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcppkg "github.com/mark3labs/mcp-go/mcp"
)

// refsFixture builds a tiny repo whose files contain calls, imports, and
// subclasses so find_references has something to find.
func refsFixture(t *testing.T) *mcpclient.Client {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"lib.py": "def helper():\n    return 1\n\nclass Base:\n    pass\n",
		"app.py": `from lib import helper, Base

class App(Base):
    def run(self):
        helper()
        helper()
`,
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

	srv := New(s, root)
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
	return c
}

func TestFindReferencesAllKinds(t *testing.T) {
	c := refsFixture(t)
	out := callJSON(t, c, "find_references", map[string]any{"name": "helper"})
	refs, _ := out["references"].([]any)
	// Expect 1 import + 2 calls = 3 refs, all in app.py.
	if len(refs) != 3 {
		t.Fatalf("expected 3 refs to helper, got %d: %v", len(refs), refs)
	}
	kinds := map[string]int{}
	for _, r := range refs {
		m := r.(map[string]any)
		kinds[m["context_kind"].(string)]++
		if m["path"] != "app.py" {
			t.Errorf("unexpected path: %v", m["path"])
		}
	}
	if kinds["call"] != 2 || kinds["import"] != 1 {
		t.Errorf("kind breakdown wrong: %v", kinds)
	}
}

func TestFindReferencesKindFilter(t *testing.T) {
	c := refsFixture(t)
	out := callJSON(t, c, "find_references", map[string]any{
		"name": "Base", "kinds": []any{"subclass"},
	})
	refs, _ := out["references"].([]any)
	if len(refs) != 1 {
		t.Fatalf("expected exactly 1 subclass ref to Base, got %d: %v", len(refs), refs)
	}
	first := refs[0].(map[string]any)
	if first["context_kind"] != "subclass" {
		t.Errorf("context_kind = %v, want subclass", first["context_kind"])
	}
}

func TestFindReferencesMissingNameErrors(t *testing.T) {
	c := refsFixture(t)
	req := mcppkg.CallToolRequest{}
	req.Params.Name = "find_references"
	req.Params.Arguments = map[string]any{}
	res, err := c.CallTool(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Errorf("expected error result for missing name")
	}
}
