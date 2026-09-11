package ids_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
)

func TestParse_Valid(t *testing.T) {
	got, err := ids.Parse(ids.Spec, "SPEC-014", 3)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	want := ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 14, Width: 3}
	if got != want {
		t.Errorf("Parse() = %+v, want %+v", got, want)
	}
	if got.String() != "SPEC-014" {
		t.Errorf("String() = %q, want %q", got.String(), "SPEC-014")
	}
}

func TestParse_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		typ   ids.EntityType
		raw   string
		width int
	}{
		{"wrong prefix", ids.Spec, "FEAT-014", 3},
		{"non-numeric suffix", ids.Spec, "SPEC-abc", 3},
		{"width too short", ids.Spec, "SPEC-14", 3},
		{"width too long", ids.Spec, "SPEC-0014", 3},
		{"missing separator", ids.Spec, "SPEC014", 3},
		{"empty", ids.Spec, "", 3},
		{"zero number", ids.Spec, "SPEC-000", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ids.Parse(tt.typ, tt.raw, tt.width)
			if !errors.Is(err, ids.ErrInvalidIDSyntax) {
				t.Fatalf("Parse(%q) error = %v, want errors.Is(err, ErrInvalidIDSyntax)", tt.raw, err)
			}
		})
	}
}

func TestNextID_EmptySlice(t *testing.T) {
	got := ids.NextID(nil, ids.Spec, 3)
	want := ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 1, Width: 3}
	if got != want {
		t.Errorf("NextID(nil) = %+v, want %+v", got, want)
	}
}

func TestNextID_GapsNotBackfilled(t *testing.T) {
	existing := []ids.EntityID{
		{Type: ids.Spec, Prefix: "SPEC", Number: 1, Width: 3},
		{Type: ids.Spec, Prefix: "SPEC", Number: 2, Width: 3},
		{Type: ids.Spec, Prefix: "SPEC", Number: 4, Width: 3},
	}
	got := ids.NextID(existing, ids.Spec, 3)
	want := ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 5, Width: 3}
	if got != want {
		t.Errorf("NextID() = %+v, want %+v (max-plus-one, gap at 3 not backfilled)", got, want)
	}
}
