package contextengine

import (
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/context/index"
)

func TestPayloadTokens_RenderedUsesInjectedEstimator(t *testing.T) {
	rendered := "Some rendered Markdown context pack body."

	got := PayloadTokens(0, nil, rendered, stubEstimator{fixed: 42})
	if got != 42 {
		t.Errorf("PayloadTokens() = %d, want 42 (from the injected estimator, not artifacts.EstimateTokens)", got)
	}

	want := artifacts.EstimateTokens(rendered)
	got = PayloadTokens(0, nil, rendered, artifacts.DefaultEstimator{})
	if got != want {
		t.Errorf("PayloadTokens() with DefaultEstimator = %d, want %d (matching artifacts.EstimateTokens on the actually-rendered text — FR-004)", got, want)
	}
}

func TestPayloadTokens_PackageOverheadUsesInjectedEstimator(t *testing.T) {
	items := []PackageItem{
		{Path: "ai/specs/SPEC-014/spec.md", Heading: "Requirements", Fingerprint: "sha256:abc"},
	}

	got := PayloadTokens(10, items, "", stubEstimator{fixed: 3})
	if got != 10+3 {
		t.Errorf("PayloadTokens() = %d, want %d (contentTokens + one stub-estimated overhead call)", got, 10+3)
	}
}

func TestPayloadTokens_CorrespondsToActualDeliveredText(t *testing.T) {
	rendered := strings.Repeat("word ", 200)
	got := PayloadTokens(0, nil, rendered, artifacts.DefaultEstimator{})
	want := artifacts.DefaultEstimator{}.Estimate(rendered)
	if got != want {
		t.Errorf("PayloadTokens() = %d, want %d — must correspond to the text actually delivered, not a disconnected approximation (FR-004)", got, want)
	}
}

// baseConfigIdentity is a fixed, fully-populated ConfigIdentity used as
// the starting point for 043-incremental-context-reuse T002's
// one-field-at-a-time difference cases.
func baseConfigIdentity() ConfigIdentity {
	return ConfigIdentity{
		Target:               "SPEC-014",
		Intent:               "implementation",
		Task:                 "add refresh rotation",
		Query:                "refresh token",
		QueryMode:            "free",
		Budget:               6000,
		HardLimit:            12000,
		PreferSection:        false,
		RankingVersion:       1,
		ContextSchemaVersion: 6,
		Estimator:            "default",
	}
}

// TestComputeConfigHash_IdenticalInputsSameHash is
// 043-incremental-context-reuse T002 (Foundational): two identical
// ConfigIdentity values hash identically.
func TestComputeConfigHash_IdenticalInputsSameHash(t *testing.T) {
	a := baseConfigIdentity()
	b := baseConfigIdentity()
	if ComputeConfigHash(a) != ComputeConfigHash(b) {
		t.Errorf("ComputeConfigHash() differs for two identical ConfigIdentity values, want identical hashes")
	}
}

// TestComputeConfigHash_AnyFieldDifferenceChangesHash is
// 043-incremental-context-reuse T002 (Foundational): any single
// differing ConfigIdentity field changes ComputeConfigHash's own
// result (data-model.md "ConfigIdentity", contracts §1).
func TestComputeConfigHash_AnyFieldDifferenceChangesHash(t *testing.T) {
	base := baseConfigIdentity()
	baseHash := ComputeConfigHash(base)

	mutations := []func(*ConfigIdentity){
		func(c *ConfigIdentity) { c.Target = "SPEC-020" },
		func(c *ConfigIdentity) { c.Intent = "planning" },
		func(c *ConfigIdentity) { c.Task = "different task" },
		func(c *ConfigIdentity) { c.Query = "different query" },
		func(c *ConfigIdentity) { c.QueryMode = "advanced" },
		func(c *ConfigIdentity) { c.Budget = 100 },
		func(c *ConfigIdentity) { c.HardLimit = 200 },
		func(c *ConfigIdentity) { c.PreferSection = true },
		func(c *ConfigIdentity) { c.RankingVersion = 2 },
		func(c *ConfigIdentity) { c.ContextSchemaVersion = 7 },
		func(c *ConfigIdentity) { c.Estimator = "other" },
	}

	for i, mutate := range mutations {
		mutated := baseConfigIdentity()
		mutate(&mutated)
		if ComputeConfigHash(mutated) == baseHash {
			t.Errorf("mutation #%d: ComputeConfigHash() unchanged after mutating one field, want a different hash", i)
		}
	}
}

func baseItemIdentities() []PackItemIdentity {
	return []PackItemIdentity{
		{Path: "ai/.../SPEC-014/spec.md", StartLine: 10, EndLine: 20, Fingerprint: "sha256:aaa"},
		{Path: "ai/.../SPEC-014/spec.md", StartLine: 30, EndLine: 40, Fingerprint: "sha256:bbb"},
	}
}

// TestComputePackID_IdenticalInputsSameID is 043-incremental-context-
// reuse T003.
func TestComputePackID_IdenticalInputsSameID(t *testing.T) {
	a := PackIdentity{Config: baseConfigIdentity(), Items: baseItemIdentities()}
	b := PackIdentity{Config: baseConfigIdentity(), Items: baseItemIdentities()}
	if ComputePackID(a) != ComputePackID(b) {
		t.Errorf("ComputePackID() differs for two identical PackIdentity values, want identical IDs")
	}
}

// TestComputePackID_ConfigDifferenceChangesID is 043-incremental-
// context-reuse T003.
func TestComputePackID_ConfigDifferenceChangesID(t *testing.T) {
	items := baseItemIdentities()
	a := PackIdentity{Config: baseConfigIdentity(), Items: items}
	differentConfig := baseConfigIdentity()
	differentConfig.Budget = 999
	b := PackIdentity{Config: differentConfig, Items: items}
	if ComputePackID(a) == ComputePackID(b) {
		t.Error("ComputePackID() unchanged when Config differs (same Items), want a different ID")
	}
}

// TestComputePackID_ItemFingerprintDifferenceChangesID is
// 043-incremental-context-reuse T003: same identity (Path/StartLine/
// EndLine), different Fingerprint alone still changes PackID.
func TestComputePackID_ItemFingerprintDifferenceChangesID(t *testing.T) {
	cfg := baseConfigIdentity()
	a := PackIdentity{Config: cfg, Items: baseItemIdentities()}

	mutatedItems := baseItemIdentities()
	mutatedItems[0].Fingerprint = "sha256:ccc"
	b := PackIdentity{Config: cfg, Items: mutatedItems}

	if ComputePackID(a) == ComputePackID(b) {
		t.Error("ComputePackID() unchanged when only one item's Fingerprint differs, want a different ID")
	}
}

// TestComputePackID_ItemAdditionOrRemovalChangesID is
// 043-incremental-context-reuse T003.
func TestComputePackID_ItemAdditionOrRemovalChangesID(t *testing.T) {
	cfg := baseConfigIdentity()
	full := PackIdentity{Config: cfg, Items: baseItemIdentities()}
	fewer := PackIdentity{Config: cfg, Items: baseItemIdentities()[:1]}
	if ComputePackID(full) == ComputePackID(fewer) {
		t.Error("ComputePackID() unchanged when the item count differs, want a different ID")
	}
}

func baseStoredItems() []index.StoredPackItem {
	return []index.StoredPackItem{
		{Path: "ai/.../SPEC-014/spec.md", StartLine: 10, EndLine: 20, Fingerprint: "sha256:aaa", Content: "### R1 — original"},
		{Path: "ai/.../SPEC-014/spec.md", StartLine: 30, EndLine: 40, Fingerprint: "sha256:bbb", Content: "### R2 — original"},
	}
}

func basePackageItems() []PackageItem {
	return []PackageItem{
		{Path: "ai/.../SPEC-014/spec.md", Content: "### R1 — original", Fingerprint: "sha256:aaa", Location: ItemLocation{Path: "ai/.../SPEC-014/spec.md", StartLine: 10, EndLine: 20}},
		{Path: "ai/.../SPEC-014/spec.md", Content: "### R2 — original", Fingerprint: "sha256:bbb", Location: ItemLocation{Path: "ai/.../SPEC-014/spec.md", StartLine: 30, EndLine: 40}},
	}
}

// TestDiffAgainstBase_NoChangeIsAllReuse is 043-incremental-context-
// reuse T009 (US1): identical current/base produces an all-"reuse"
// diff, with no content anywhere.
func TestDiffAgainstBase_NoChangeIsAllReuse(t *testing.T) {
	diff := DiffAgainstBase(basePackageItems(), baseStoredItems())

	if len(diff.Entries) != 2 {
		t.Fatalf("Entries = %+v, want 2", diff.Entries)
	}
	for i, e := range diff.Entries {
		if e.Kind != "reuse" {
			t.Errorf("Entries[%d].Kind = %q, want %q", i, e.Kind, "reuse")
		}
		if e.Item != nil {
			t.Errorf("Entries[%d].Item = %+v, want nil (no content resent for a reuse entry)", i, e.Item)
		}
		if e.BaseIndex == nil || *e.BaseIndex != i {
			t.Errorf("Entries[%d].BaseIndex = %v, want pointer to %d", i, e.BaseIndex, i)
		}
	}
	if len(diff.Removed) != 0 {
		t.Errorf("Removed = %+v, want none", diff.Removed)
	}
}

// TestDiffAgainstBase_OneModifiedItem is 043-incremental-context-reuse
// T009 (US1): a single item whose identity is unchanged but whose
// Fingerprint differs is reported as exactly one "modified" entry;
// every other item stays "reuse".
func TestDiffAgainstBase_OneModifiedItem(t *testing.T) {
	current := basePackageItems()
	current[0].Content = "### R1 — CHANGED"
	current[0].Fingerprint = "sha256:ccc"

	diff := DiffAgainstBase(current, baseStoredItems())

	if diff.Entries[0].Kind != "modified" {
		t.Errorf("Entries[0].Kind = %q, want %q", diff.Entries[0].Kind, "modified")
	}
	if diff.Entries[0].Item == nil || diff.Entries[0].Item.Content != "### R1 — CHANGED" {
		t.Errorf("Entries[0].Item = %+v, want the current item's own full data", diff.Entries[0].Item)
	}
	if diff.Entries[1].Kind != "reuse" {
		t.Errorf("Entries[1].Kind = %q, want %q (unaffected item)", diff.Entries[1].Kind, "reuse")
	}
}

// TestDiffAgainstBase_AddedItem is 043-incremental-context-reuse T009
// (US1): an item whose identity is new to current is reported as
// "added".
func TestDiffAgainstBase_AddedItem(t *testing.T) {
	current := basePackageItems()
	current = append(current, PackageItem{
		Path: "ai/.../SPEC-014/spec.md", Content: "### R3 — new", Fingerprint: "sha256:ddd",
		Location: ItemLocation{Path: "ai/.../SPEC-014/spec.md", StartLine: 50, EndLine: 60},
	})

	diff := DiffAgainstBase(current, baseStoredItems())

	if len(diff.Entries) != 3 {
		t.Fatalf("Entries = %+v, want 3", diff.Entries)
	}
	if diff.Entries[2].Kind != "added" || diff.Entries[2].Item == nil || diff.Entries[2].Item.Content != "### R3 — new" {
		t.Errorf("Entries[2] = %+v, want an added entry carrying the new item", diff.Entries[2])
	}
}

// TestDiffAgainstBase_RemovedItem is 043-incremental-context-reuse
// T009 (US1): a base identity absent from current is reported under
// Removed.
func TestDiffAgainstBase_RemovedItem(t *testing.T) {
	current := basePackageItems()[:1] // drop the second item

	diff := DiffAgainstBase(current, baseStoredItems())

	if len(diff.Removed) != 1 {
		t.Fatalf("Removed = %+v, want 1 entry", diff.Removed)
	}
	if diff.Removed[0].StartLine != 30 || diff.Removed[0].EndLine != 40 {
		t.Errorf("Removed[0] = %+v, want the second base item's own identity", diff.Removed[0])
	}
	if len(diff.Entries) != 1 || diff.Entries[0].Kind != "reuse" {
		t.Errorf("Entries = %+v, want exactly one reuse entry for the surviving item", diff.Entries)
	}
}

// TestDiffAgainstBase_ReconstructsExactFullPack is
// 043-incremental-context-reuse T010 (US1, spec FR-003): applying a
// PackDiff (substituting each "reuse" entry with base[BaseIndex],
// keeping every "added"/"modified" entry's own Item, in Entries'
// own order) reproduces current exactly, item by item.
func TestDiffAgainstBase_ReconstructsExactFullPack(t *testing.T) {
	base := baseStoredItems()
	current := basePackageItems()
	current[0].Content = "### R1 — CHANGED"
	current[0].Fingerprint = "sha256:ccc"
	current = append(current, PackageItem{
		Path: "ai/.../SPEC-014/spec.md", Content: "### R3 — new", Fingerprint: "sha256:ddd",
		Location: ItemLocation{Path: "ai/.../SPEC-014/spec.md", StartLine: 50, EndLine: 60},
	})

	diff := DiffAgainstBase(current, base)

	reconstructed := make([]PackageItem, 0, len(diff.Entries))
	for _, e := range diff.Entries {
		switch e.Kind {
		case "reuse":
			b := base[*e.BaseIndex]
			reconstructed = append(reconstructed, PackageItem{
				Path:        b.Path,
				Content:     b.Content,
				Fingerprint: b.Fingerprint,
				Location:    ItemLocation{Path: b.Path, StartLine: b.StartLine, EndLine: b.EndLine},
			})
		case "added", "modified":
			reconstructed = append(reconstructed, *e.Item)
		}
	}

	if len(reconstructed) != len(current) {
		t.Fatalf("reconstructed %d items, want %d", len(reconstructed), len(current))
	}
	for i := range current {
		if reconstructed[i].Content != current[i].Content || reconstructed[i].Fingerprint != current[i].Fingerprint {
			t.Errorf("reconstructed[%d] = %+v, want %+v", i, reconstructed[i], current[i])
		}
	}
}
