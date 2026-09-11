package agents

import (
	"context"
	"io/fs"

	"github.com/mottamarcio/misterspec/internal/installer"
)

// Adapter is one coding-agent integration
// (docs/architecture-specification.md §34). An adapter may identify its
// integration directory, transform canonical Skills where necessary,
// copy Skill files, and create agent-specific metadata — it may never
// alter what a Skill means, modify project-specific requirements, or
// become an authoritative source of project state (§35, FR-008).
type Adapter interface {
	// ID is this adapter's stable identifier (e.g. "claude-code").
	ID() string
	// Name is this adapter's human-readable name (e.g. "Claude Code").
	Name() string
	// TargetPath is this adapter's integration location, relative to a
	// project root (e.g. ".claude/skills") — discoverable without
	// installing anything (FR-009).
	TargetPath() string
	// Install materializes req.Skills into TargetPath() and records what
	// happened — one call, one complete outcome (research.md).
	//
	// ctx is kept per docs/architecture-specification.md §34's frozen
	// interface shape, even though this feature's local-filesystem-only
	// work has nothing cancellable to do with it yet (research.md).
	Install(ctx context.Context, req InstallRequest) (InstallResult, error)
}

// InstallRequest is the input to Adapter.Install.
type InstallRequest struct {
	// ProjectRoot is the target project's root directory.
	ProjectRoot string
	// Skills is the source of Skill resources to materialize — a
	// fixture in this feature's own tests; kit.SkillsFS once a later
	// feature (Phase 6) gives it real content (research.md).
	Skills fs.FS
	// Overwrite has the same meaning as installer.Install's overwrite.
	Overwrite bool
}

// InstallResult is the in-memory outcome of one Adapter.Install call
// (FR-006).
type InstallResult struct {
	// AdapterID is which adapter ran.
	AdapterID string
	// IntegrationPath is where it installed to, relative to ProjectRoot.
	IntegrationPath string
	// Outcomes is the per-resource result, reused directly from
	// internal/installer — not a second, parallel result type.
	Outcomes []installer.Outcome
}
