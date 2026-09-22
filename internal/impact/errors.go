package impact

import "errors"

// ErrRevisionNotFound is returned when --from or a non-empty --to does
// not resolve in the project's Git repository (contracts §4).
var ErrRevisionNotFound = errors.New("impact: revision not found")

// ErrNotARepository is returned when the target project directory is
// not a Git repository at all — AnalyzeImpact cannot run without Git
// history to diff (contracts §4).
var ErrNotARepository = errors.New("impact: not a git repository")
