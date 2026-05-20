package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPreflightFlagsStaleBinary(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "fakebin")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho 0.0.1"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Backdate the binary so it's older than any source tree we walk.
	old := time.Now().Add(-30 * 24 * time.Hour)
	if err := os.Chtimes(bin, old, old); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "mcp.json")
	body := `{"mcpServers":{"cogni":{"command":"` + bin + `","args":["serve"]}}}`
	if err := os.WriteFile(cfg, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := preflightMCPBinaries(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "STALE") {
		t.Errorf("expected STALE marker, got:\n%s", out)
	}
	if !strings.Contains(out, "Rebuild with") {
		t.Errorf("expected rebuild hint, got:\n%s", out)
	}
}

func TestPreflightSkipsWhenNoConfig(t *testing.T) {
	var buf bytes.Buffer
	if err := preflightMCPBinaries(&buf, ""); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output when no config, got %q", buf.String())
	}
}

func TestPreflightReportsMissingBinary(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "mcp.json")
	body := `{"mcpServers":{"cogni":{"command":"/does/not/exist","args":[]}}}`
	if err := os.WriteFile(cfg, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := preflightMCPBinaries(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "[ERR]") {
		t.Errorf("expected [ERR] marker for missing binary, got:\n%s", buf.String())
	}
}
