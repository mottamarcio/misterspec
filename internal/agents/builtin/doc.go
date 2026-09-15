// Package builtin wires together every concrete internal/agents.Adapter
// this build knows about into a default Registry. It exists as its own
// package specifically to avoid the import cycle a naive design (package
// agents both defining Adapter and constructing concrete adapters like
// claude.New()) would create — see
// specs/006-agent-adapter/research.md.
package builtin
