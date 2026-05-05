// Command cogni is the Cogni MCP server CLI.
//
// It exposes three subcommands: `serve` (run the MCP server over stdio),
// `index` (build or refresh the on-disk index for a repository), and
// `bench` (run the token-savings benchmark harness).
package main

func main() {
	Execute()
}
