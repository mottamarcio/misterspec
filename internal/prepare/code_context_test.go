package prepare

import (
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/context/index"
)

// stubCodeStore is a minimal index.Store stub exercising only
// DeclarationsForFiles — every other method panics if called, since
// ResolveCodeContext has no reason to call anything else.
type stubCodeStore struct {
	index.Store
	decls map[string][]index.CodeDeclaration
}

func (s stubCodeStore) DeclarationsForFiles(paths []string) ([]index.CodeDeclaration, error) {
	var out []index.CodeDeclaration
	for _, p := range paths {
		out = append(out, s.decls[p]...)
	}
	return out, nil
}

// TestResolveCodeContext_SignatureBySizeThreshold is
// 044-architecture-code-context-rules T022 (US3): a small file's own
// declaration is rendered at full Body; a large file's own declaration
// is rendered at Signature only (research.md #7).
func TestResolveCodeContext_SignatureBySizeThreshold(t *testing.T) {
	small := index.CodeDeclaration{
		Path: "internal/foo/small.go", Name: "Small", Kind: "func",
		Signature: "func Small()", Body: "func Small() { return }",
		StartLine: 3, EndLine: 3,
	}
	large := index.CodeDeclaration{
		Path: "internal/foo/large.go", Name: "Large", Kind: "func",
		Signature: "func Large()", Body: "func Large() {\n" + strings.Repeat("doSomething()\n", 2000) + "}",
		StartLine: 3, EndLine: 2003,
	}
	store := stubCodeStore{decls: map[string][]index.CodeDeclaration{
		"internal/foo/small.go": {small},
		"internal/foo/large.go": {large},
	}}

	items, notFound, err := ResolveCodeContext(store, "internal/foo/small.go, internal/foo/large.go", artifacts.DefaultEstimator{})
	if err != nil {
		t.Fatalf("ResolveCodeContext() unexpected error: %v", err)
	}
	if len(notFound) != 0 {
		t.Errorf("notFound = %+v, want none", notFound)
	}
	if len(items) != 2 {
		t.Fatalf("items = %+v, want 2", items)
	}

	byHeading := map[string]string{}
	for _, item := range items {
		byHeading[item.Heading] = item.Content
	}
	if byHeading["Small"] != small.Body {
		t.Errorf("Small.Content = %q, want the full Body (small file, under threshold)", byHeading["Small"])
	}
	if byHeading["Large"] != large.Signature {
		t.Errorf("Large.Content = %q, want the Signature only (large file, over threshold)", byHeading["Large"])
	}
}

// TestResolveCodeContext_UnresolvedScopePathReported is
// 044-architecture-code-context-rules T022 (US3, spec Edge Case): a
// Scope path with no indexed match is returned in notFound, not
// silently dropped.
func TestResolveCodeContext_UnresolvedScopePathReported(t *testing.T) {
	store := stubCodeStore{decls: map[string][]index.CodeDeclaration{}}

	items, notFound, err := ResolveCodeContext(store, "internal/gone/deleted.go", artifacts.DefaultEstimator{})
	if err != nil {
		t.Fatalf("ResolveCodeContext() unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("items = %+v, want none", items)
	}
	if len(notFound) != 1 || notFound[0] != "internal/gone/deleted.go" {
		t.Errorf("notFound = %+v, want [\"internal/gone/deleted.go\"]", notFound)
	}
}

// TestResolveCodeContext_EmptyScopeIsEmpty is
// 044-architecture-code-context-rules T022: an empty Scope returns no
// items and no notFound entries.
func TestResolveCodeContext_EmptyScopeIsEmpty(t *testing.T) {
	store := stubCodeStore{decls: map[string][]index.CodeDeclaration{}}

	items, notFound, err := ResolveCodeContext(store, "", artifacts.DefaultEstimator{})
	if err != nil {
		t.Fatalf("ResolveCodeContext() unexpected error: %v", err)
	}
	if len(items) != 0 || len(notFound) != 0 {
		t.Errorf("items/notFound = %+v/%+v, want both empty", items, notFound)
	}
}
