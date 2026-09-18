package ids_test

import (
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
)

func taskRefTestConfig() project.Configuration {
	return project.Configuration{IDWidth: 3}
}

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

func TestTaskID_StringAndEquality(t *testing.T) {
	spec := ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 14, Width: 3}
	local := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 3, Width: 3}
	id := ids.TaskID{Spec: spec, Local: local}

	if got, want := id.String(), "SPEC-014:TASK-003"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	same := ids.TaskID{Spec: spec, Local: local}
	if id != same {
		t.Errorf("TaskID{%v} != TaskID{%v}, want equal", id, same)
	}

	otherSpec := ids.TaskID{Spec: ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 15, Width: 3}, Local: local}
	if id == otherSpec {
		t.Errorf("TaskID with different owning Specs compared equal: %v == %v", id, otherSpec)
	}
}

func TestParseTaskRef_ValidComposite(t *testing.T) {
	cfg := taskRefTestConfig()
	got, err := ids.ParseTaskRef("SPEC-014:TASK-003", cfg)
	if err != nil {
		t.Fatalf("ParseTaskRef() unexpected error: %v", err)
	}
	want := ids.TaskID{
		Spec:  ids.EntityID{Type: ids.Spec, Prefix: "SPEC", Number: 14, Width: 3},
		Local: ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 3, Width: 3},
	}
	if got != want {
		t.Errorf("ParseTaskRef() = %+v, want %+v", got, want)
	}
}

func TestParseTaskRef_ValidBare(t *testing.T) {
	cfg := taskRefTestConfig()
	got, err := ids.ParseTaskRef("TASK-003", cfg)
	if err != nil {
		t.Fatalf("ParseTaskRef() unexpected error: %v", err)
	}
	if got.Spec.Number != 0 {
		t.Errorf("ParseTaskRef() bare form Spec = %+v, want zero-value (no Spec half given)", got.Spec)
	}
	want := ids.EntityID{Type: ids.Task, Prefix: "TASK", Number: 3, Width: 3}
	if got.Local != want {
		t.Errorf("ParseTaskRef() Local = %+v, want %+v", got.Local, want)
	}
}

func TestParseTaskRef_Invalid(t *testing.T) {
	cfg := taskRefTestConfig()
	tests := []struct {
		name string
		raw  string
	}{
		{"missing task half", "SPEC-014:"},
		{"missing spec half", ":TASK-003"},
		{"wrong separator", "SPEC-014-TASK-003"},
		{"wrong ID width on spec half", "SPEC-14:TASK-003"},
		{"wrong ID width on task half", "SPEC-014:TASK-3"},
		{"wrong entity type in spec slot", "TASK-003:TASK-004"},
		{"wrong entity type in task slot", "SPEC-014:SPEC-003"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ids.ParseTaskRef(tt.raw, cfg)
			if !errors.Is(err, ids.ErrInvalidIDSyntax) {
				t.Fatalf("ParseTaskRef(%q) error = %v, want errors.Is(err, ErrInvalidIDSyntax)", tt.raw, err)
			}
		})
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
