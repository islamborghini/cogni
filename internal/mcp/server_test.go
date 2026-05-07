package mcp

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/islamborghini/cogni/internal/store"
	mcpclient "github.com/mark3labs/mcp-go/client"
	mcppkg "github.com/mark3labs/mcp-go/mcp"
)

func TestServerRegistersFiveTools(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	srv := New(s, t.TempDir())
	c, err := mcpclient.NewInProcessClient(srv.MCP())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatal(err)
	}
	init := mcppkg.InitializeRequest{}
	init.Params.ProtocolVersion = mcppkg.LATEST_PROTOCOL_VERSION
	init.Params.ClientInfo = mcppkg.Implementation{Name: "test", Version: "0.0.1"}
	if _, err := c.Initialize(ctx, init); err != nil {
		t.Fatal(err)
	}

	list, err := c.ListTools(ctx, mcppkg.ListToolsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(list.Tools))
	for _, tl := range list.Tools {
		got = append(got, tl.Name)
	}
	sort.Strings(got)
	want := []string{"file_outline", "find_references", "repo_overview", "symbol_search", "symbol_source"}
	if len(got) != len(want) {
		t.Fatalf("expected 5 tools, got %d: %v", len(got), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("tool[%d]: want %q, got %q", i, w, got[i])
		}
	}
}
