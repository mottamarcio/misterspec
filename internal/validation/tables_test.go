package validation

import (
	"reflect"
	"sort"
	"testing"

	"github.com/mottamarcio/misterspec/internal/ids"
)

func sortedCopy(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

func TestAllowedStates(t *testing.T) {
	tests := []struct {
		name    string
		typ     ids.EntityType
		want    []string
		wantHas bool
	}{
		{"program", ids.Program, []string{"draft", "active", "done", "cancelled"}, true},
		{"feature", ids.Feature, []string{"draft", "active", "done", "cancelled"}, true},
		{"spec", ids.Spec, []string{"draft", "ready", "in_progress", "validated", "blocked", "superseded", "cancelled"}, true},
		{"learning", ids.Learning, []string{"candidate", "promoted", "dismissed"}, true},
		{"knowledge", ids.Knowledge, []string{"active"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := allowedStates(tt.typ)
			if ok != tt.wantHas {
				t.Fatalf("allowedStates(%v) ok = %v, want %v", tt.typ, ok, tt.wantHas)
			}
			if !reflect.DeepEqual(sortedCopy(got), sortedCopy(tt.want)) {
				t.Errorf("allowedStates(%v) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}

func TestRequiredParentType(t *testing.T) {
	tests := []struct {
		name    string
		typ     ids.EntityType
		want    ids.EntityType
		wantHas bool
	}{
		{"feature requires program", ids.Feature, ids.Program, true},
		{"spec requires feature", ids.Spec, ids.Feature, true},
		{"program has no parent", ids.Program, 0, false},
		{"knowledge has no parent", ids.Knowledge, 0, false},
		{"learning has no parent", ids.Learning, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := requiredParentType(tt.typ)
			if ok != tt.wantHas {
				t.Fatalf("requiredParentType(%v) ok = %v, want %v", tt.typ, ok, tt.wantHas)
			}
			if ok && got != tt.want {
				t.Errorf("requiredParentType(%v) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}
