package internalcmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/validation"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name         string
		err          error
		wantCode     string
		wantExitCode int
	}{
		{"not initialized", project.ErrNotInitialized, "project_not_initialized", 6},
		{"entity not found", operations.ErrEntityNotFound, "entity_not_found", 3},
		{"entity ambiguous", &operations.AmbiguousIDError{ID: "SPEC-001", Locations: []string{"a", "b"}}, "entity_ambiguous", 3},
		{"invalid target (operations)", operations.ErrInvalidTarget, "invalid_target", 2},
		{"invalid target (validation)", validation.ErrInvalidTarget, "invalid_target", 2},
		{"invalid parent", operations.ErrInvalidParent, "invalid_parent", 5},
		{"already exists", operations.ErrAlreadyExists, "already_exists", 5},
		{"unsupported type", operations.ErrUnsupportedType, "unsupported_type", 2},
		{"invalid slug", operations.ErrInvalidSlug, "invalid_argument", 2},
		{"cli invalid argument", ErrInvalidArgument, "invalid_argument", 2},
		{"path outside project", artifacts.ErrPathOutsideProject, "path_outside_project", 5},
		{"already initialized", bootstrap.ErrAlreadyInitialized, "already_initialized", 5},
		{"unknown agent", bootstrap.ErrUnknownAgent, "unknown_agent", 5},
		{"unrecognized fallback", errors.New("boom"), "unexpected_failure", 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, exitCode := classify(tc.err)
			if code != tc.wantCode {
				t.Errorf("classify(%v) code = %q, want %q", tc.err, code, tc.wantCode)
			}
			if exitCode != tc.wantExitCode {
				t.Errorf("classify(%v) exitCode = %d, want %d", tc.err, exitCode, tc.wantExitCode)
			}
		})
	}
}

func TestClassify_WrappedError(t *testing.T) {
	wrapped := fmt.Errorf("resolving %s: %w", "SPEC-014", operations.ErrEntityNotFound)

	code, exitCode := classify(wrapped)
	if code != "entity_not_found" || exitCode != 3 {
		t.Errorf("classify(wrapped ErrEntityNotFound) = (%q, %d), want (\"entity_not_found\", 3)", code, exitCode)
	}
}
