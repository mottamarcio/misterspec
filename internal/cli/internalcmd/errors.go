package internalcmd

import (
	"errors"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// ErrInvalidArgument is internalcmd's own sentinel for a CLI-boundary
// problem detected before any operations/validation/bootstrap function
// is even called — an unrecognized --type/kind name, a missing required
// flag, or a Cobra-level parsing failure (data-model.md's Error Code
// table; distinct from operations.ErrInvalidSlug, which classify also
// maps to the same "invalid_argument" code).
var ErrInvalidArgument = errors.New("internalcmd: invalid argument")

// classify maps err to (code, exitCode) per data-model.md's Error Code
// table — the single place every sentinel-to-code mapping lives
// (research.md), so a new sentinel is wired in once here rather than
// reimplemented per command. An unrecognized error classifies as
// ("unexpected_failure", 1).
func classify(err error) (code string, exitCode int) {
	switch {
	case errors.Is(err, project.ErrNotInitialized):
		return "project_not_initialized", 6
	case errors.Is(err, operations.ErrEntityNotFound):
		return "entity_not_found", 3
	case errors.Is(err, operations.ErrEntityAmbiguous):
		return "entity_ambiguous", 3
	case errors.Is(err, operations.ErrInvalidTarget), errors.Is(err, validation.ErrInvalidTarget):
		return "invalid_target", 2
	case errors.Is(err, operations.ErrInvalidParent):
		return "invalid_parent", 5
	case errors.Is(err, operations.ErrAlreadyExists):
		return "already_exists", 5
	case errors.Is(err, operations.ErrUnsupportedType):
		return "unsupported_type", 2
	case errors.Is(err, operations.ErrInvalidSlug), errors.Is(err, ErrInvalidArgument):
		return "invalid_argument", 2
	case errors.Is(err, contextengine.ErrUnsupportedIntent):
		return "unsupported_intent", 2
	case errors.Is(err, artifacts.ErrPathOutsideProject):
		return "path_outside_project", 5
	case errors.Is(err, bootstrap.ErrAlreadyInitialized):
		return "already_initialized", 5
	case errors.Is(err, bootstrap.ErrUnknownAgent):
		return "unknown_agent", 5
	default:
		return "unexpected_failure", 1
	}
}
