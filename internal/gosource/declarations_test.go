package gosource

import (
	"strings"
	"testing"
)

// TestDeclarations_FuncTypeConstVar is 044-architecture-code-context-
// rules T004 (Foundational): Declarations returns one Declaration per
// top-level func/type/const/var, each with Name/Kind/Signature (with
// doc comment)/Body/StartLine/EndLine (data-model.md "CodeDeclaration",
// contracts §1).
func TestDeclarations_FuncTypeConstVar(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "decls.go", `package foo

// Bar does a thing.
func Bar(x int) error {
	return nil
}

// Baz is a type.
type Baz struct {
	X int
}

const Qux = 1

var Quux = "hello"
`)

	got, err := Declarations(path)
	if err != nil {
		t.Fatalf("Declarations() unexpected error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("Declarations() = %+v, want 4 entries", got)
	}

	byName := map[string]Declaration{}
	for _, d := range got {
		byName[d.Name] = d
	}

	bar, ok := byName["Bar"]
	if !ok {
		t.Fatalf("no declaration named Bar in %+v", got)
	}
	if bar.Kind != "func" {
		t.Errorf("Bar.Kind = %q, want \"func\"", bar.Kind)
	}
	if !strings.Contains(bar.Signature, "Bar does a thing.") {
		t.Errorf("Bar.Signature = %q, want it to contain the doc comment", bar.Signature)
	}
	if !strings.Contains(bar.Signature, "func Bar(x int) error") {
		t.Errorf("Bar.Signature = %q, want it to contain the func signature", bar.Signature)
	}
	if !strings.Contains(bar.Body, "return nil") {
		t.Errorf("Bar.Body = %q, want it to contain the full body", bar.Body)
	}
	if bar.StartLine == 0 || bar.EndLine == 0 || bar.EndLine < bar.StartLine {
		t.Errorf("Bar StartLine/EndLine = %d/%d, want a real, ordered range", bar.StartLine, bar.EndLine)
	}

	baz, ok := byName["Baz"]
	if !ok || baz.Kind != "type" {
		t.Errorf("Baz declaration = %+v (ok=%v), want Kind \"type\"", baz, ok)
	}

	qux, ok := byName["Qux"]
	if !ok || qux.Kind != "const" {
		t.Errorf("Qux declaration = %+v (ok=%v), want Kind \"const\"", qux, ok)
	}

	quux, ok := byName["Quux"]
	if !ok || quux.Kind != "var" {
		t.Errorf("Quux declaration = %+v (ok=%v), want Kind \"var\"", quux, ok)
	}
}

// TestDeclarations_GroupedConstBlockHasPerDeclarationLines is a
// code-review regression: a parenthesized "const (...)" block must
// yield one Declaration per identifier, each with its own line range
// and body text — not every identifier sharing the whole block's
// Body/StartLine/EndLine (which corrupts code_context locations and
// inflates size estimates for unrelated identifiers in the same
// block).
func TestDeclarations_GroupedConstBlockHasPerDeclarationLines(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "consts.go", `package foo

const (
	StatusPass        = "pass"
	StatusFail        = "fail"
	StatusNotEvaluated = "not_evaluated"
)
`)

	got, err := Declarations(path)
	if err != nil {
		t.Fatalf("Declarations() unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("Declarations() = %+v, want 3 entries", got)
	}

	byName := map[string]Declaration{}
	for _, d := range got {
		byName[d.Name] = d
	}

	pass, fail := byName["StatusPass"], byName["StatusFail"]
	if pass.StartLine == fail.StartLine {
		t.Errorf("StatusPass and StatusFail share StartLine %d, want distinct per-identifier lines", pass.StartLine)
	}
	if strings.Contains(pass.Body, "StatusFail") {
		t.Errorf("StatusPass.Body = %q, want it to not include the sibling StatusFail declaration", pass.Body)
	}
	if !strings.Contains(pass.Body, `"pass"`) {
		t.Errorf("StatusPass.Body = %q, want it to contain its own value", pass.Body)
	}
}

// TestDeclarations_OnlyCommentsAndPackageClause is
// 044-architecture-code-context-rules T004: a file with only a
// package clause and comments returns an empty, non-nil slice.
func TestDeclarations_OnlyCommentsAndPackageClause(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "empty.go", "// Package foo does nothing.\npackage foo\n")

	got, err := Declarations(path)
	if err != nil {
		t.Fatalf("Declarations() unexpected error: %v", err)
	}
	if got == nil {
		t.Error("Declarations() = nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Declarations() = %+v, want none", got)
	}
}
