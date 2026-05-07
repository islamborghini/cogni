package mcp

import (
	"sort"
	"strings"

	"github.com/islamborghini/cogni/internal/store"
)

// PackageNode is a single entry in the repo overview tree. It represents a
// directory at depth ≤ maxDepth; deeper files are aggregated into the closest
// ancestor's FileCount.
type PackageNode struct {
	Path       string        `json:"path"`
	Name       string        `json:"name"`
	FileCount  int           `json:"file_count"`
	TopSymbols []OverviewSym `json:"top_symbols,omitempty"`
}

// OverviewSym is a slimmed-down symbol record for repo_overview.
type OverviewSym struct {
	Name      string `json:"name"`
	Qualified string `json:"qualified"`
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
}

// BuildOverview groups indexed file paths into a package tree truncated to
// maxDepth and attaches up to topPerPkg exported symbols from each package.
func BuildOverview(s *store.Store, maxDepth, topPerPkg int) ([]PackageNode, error) {
	if maxDepth < 1 {
		maxDepth = 3
	}
	if topPerPkg < 1 {
		topPerPkg = 5
	}
	paths, err := s.AllFilePaths()
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	for _, p := range paths {
		// Each file contributes to every prefix from depth 1 up to maxDepth.
		parts := strings.Split(p, "/")
		// Drop the filename — packages are directories.
		if len(parts) > 1 {
			parts = parts[:len(parts)-1]
		} else {
			parts = []string{"."}
		}
		depth := len(parts)
		if depth > maxDepth {
			depth = maxDepth
		}
		for d := 1; d <= depth; d++ {
			prefix := strings.Join(parts[:d], "/")
			counts[prefix]++
		}
	}

	prefixes := make([]string, 0, len(counts))
	for k := range counts {
		prefixes = append(prefixes, k)
	}
	sort.Strings(prefixes)

	nodes := make([]PackageNode, 0, len(prefixes))
	for _, pre := range prefixes {
		node := PackageNode{
			Path:      pre,
			Name:      lastSegment(pre),
			FileCount: counts[pre],
		}
		syms, err := s.TopExportedSymbolsByPrefix(pre, topPerPkg)
		if err != nil {
			return nil, err
		}
		for _, sm := range syms {
			node.TopSymbols = append(node.TopSymbols, OverviewSym{
				Name:      sm.Name,
				Qualified: sm.Qualified,
				Kind:      sm.Kind,
				Path:      sm.FilePath,
				Line:      sm.StartLine,
			})
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func lastSegment(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}
