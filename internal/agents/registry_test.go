package agents_test

import (
	"context"
	"testing"

	"github.com/mottamarcio/misterspec/internal/agents"
)

// fakeAdapter is a minimal agents.Adapter used only to test Registry —
// never the real claude package, so User Story 1 is provable without
// User Story 2's Install being ready (tasks.md's Organization note).
type fakeAdapter struct {
	id, name, target string
}

func (f fakeAdapter) ID() string         { return f.id }
func (f fakeAdapter) Name() string       { return f.name }
func (f fakeAdapter) TargetPath() string { return f.target }
func (f fakeAdapter) Install(ctx context.Context, req agents.InstallRequest) (agents.InstallResult, error) {
	return agents.InstallResult{AdapterID: f.id, IntegrationPath: f.target}, nil
}

func TestRegistry_ListSortedByID(t *testing.T) {
	registry := agents.NewRegistry(
		fakeAdapter{id: "zeta", name: "Zeta", target: ".zeta/skills"},
		fakeAdapter{id: "alpha", name: "Alpha", target: ".alpha/skills"},
	)

	list := registry.List()
	if len(list) != 2 {
		t.Fatalf("List() = %d adapters, want 2", len(list))
	}
	if list[0].ID() != "alpha" || list[1].ID() != "zeta" {
		t.Errorf("List() order = [%s, %s], want [alpha, zeta]", list[0].ID(), list[1].ID())
	}
}

func TestRegistry_GetRegistered(t *testing.T) {
	registry := agents.NewRegistry(fakeAdapter{id: "alpha", name: "Alpha", target: ".alpha/skills"})

	got, ok := registry.Get("alpha")
	if !ok {
		t.Fatal("Get(\"alpha\") ok = false, want true")
	}
	if got.Name() != "Alpha" || got.TargetPath() != ".alpha/skills" {
		t.Errorf("Get(\"alpha\") = %+v, unexpected", got)
	}
}

func TestRegistry_GetUnregistered(t *testing.T) {
	registry := agents.NewRegistry(fakeAdapter{id: "alpha", name: "Alpha", target: ".alpha/skills"})

	_, ok := registry.Get("does-not-exist")
	if ok {
		t.Fatal("Get(\"does-not-exist\") ok = true, want false")
	}
}
