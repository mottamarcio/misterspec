package internalcmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/context/index"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/operations"
	"github.com/mottamarcio/misterspec/internal/prepare"
	"github.com/mottamarcio/misterspec/internal/project"
)

// NewPrepareCmd builds "misterspec internal prepare SPEC-###
// [--task TASK-###]" (034-task-oriented-context-preparation
// contracts/prepare-command.md §1). It resolves the Spec, scans its own
// tasks.md once via prepare.ScanSpecTasks, and assembles one read-only
// TaskPreparation for the named or automatically selected Task — no
// command is ever executed (spec FR-002).
func NewPrepareCmd() *cobra.Command {
	var dir, task string

	cmd := &cobra.Command{
		Use:           "prepare SPEC-### [--task TASK-###]",
		Short:         "Assemble everything needed to start one Task in a single response",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := project.Detect(dir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}

			loc, err := operations.Resolve(proj.Root, proj.Config, args[0])
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			if loc.ID.Type != ids.Spec {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %s is not a Spec", operations.ErrInvalidTarget, args[0]))
			}
			spec := loc.ID
			specDir := strings.TrimSuffix(loc.Path, "/spec.md")

			infos, found, err := prepare.ScanSpecTasks(proj.Root, proj.Config, spec, specDir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			if !found {
				return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %s has no tasks.md yet", operations.ErrEntityNotFound, spec))
			}

			byNumber := map[int]prepare.TaskInfo{}
			for _, info := range infos {
				byNumber[info.Task.Number] = info
			}
			statusOf := func(id ids.EntityID) string {
				if info, ok := byNumber[id.Number]; ok {
					return info.Status
				}
				return "pending"
			}

			var target prepare.TaskInfo
			var haveTarget bool
			if task != "" {
				num, parseErr := parseTaskNumberArg(task, spec, proj.Config)
				if parseErr != nil {
					return WriteError(cmd.OutOrStdout(), parseErr)
				}
				info, ok := byNumber[num]
				if !ok {
					return WriteError(cmd.OutOrStdout(), fmt.Errorf("%w: %s:TASK-%d", operations.ErrEntityNotFound, spec, num))
				}
				target, haveTarget = info, true
			} else {
				target, haveTarget = prepare.SelectTask(infos, statusOf)
			}

			if !haveTarget {
				return WriteSuccess(cmd.OutOrStdout(), map[string]any{
					"preparation": nil,
					"message":     fmt.Sprintf("no Task in %s is ready — every Task is either complete or blocked", spec),
				})
			}

			specBody, err := readSpecBody(proj.Root, specDir)
			if err != nil {
				return WriteError(cmd.OutOrStdout(), err)
			}
			planBody := readPlanBodyBestEffort(proj.Root, specDir)

			readiness := prepare.ComputeReadinessWithCycleCheck(target.Task, target.Dependency.DependsOn, statusOf, infos)
			requirements := prepare.LookupRequirementTexts(specBody, target.Coverage.References)
			planSections := prepare.AssociatePlanSections(planBody, target.Coverage.References)

			codeContext, codeScopeNotFound := resolveCodeContextBestEffort(proj, target.Fields.Scope)

			result := prepare.BuildTaskPreparation(target.Task, target.Heading, readiness, requirements, target.Fields, planSections, codeContext, codeScopeNotFound)

			if writeErr := WriteSuccess(cmd.OutOrStdout(), map[string]any{
				"preparation": renderPreparation(result, spec),
			}); writeErr != nil {
				return writeErr
			}
			if !readiness.Ready {
				return &ExitCodeError{Code: 10}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "target project directory")
	cmd.Flags().StringVar(&task, "task", "", "the Task to prepare (bare TASK-### or composite SPEC-###:TASK-###)")
	return cmd
}

// resolveCodeContextBestEffort opens/syncs the same disposable code
// index internal context already owns (044-architecture-code-context-
// rules research.md #6), then resolves scope against it via
// prepare.ResolveCodeContext. Mirrors context.go's own
// index.Open/SyncCode lifecycle — one open per call, closed before
// returning. True to its name: any failure opening or syncing the code
// index (e.g. an unwritable cache directory) is swallowed rather than
// aborting the whole `internal prepare` response — code_context is
// additive, and a Task whose Scope names no .go file at all should
// never fail to prepare over it.
func resolveCodeContextBestEffort(proj *project.Project, scope string) ([]contextengine.PackageItem, []string) {
	cachePath := filepath.Join(proj.Root, ".misterspec", "cache", "context.db")
	store, err := index.Open(cachePath)
	if err != nil {
		return nil, nil
	}
	defer store.Close()

	if _, err := store.SyncCode(proj.Root, proj.Config.CodeExclusions); err != nil {
		return nil, nil
	}

	items, notFound, err := prepare.ResolveCodeContext(store, scope, artifacts.DefaultEstimator{})
	if err != nil {
		return nil, nil
	}
	return items, notFound
}

// readSpecBody reads the owning Spec's own spec.md body — required for
// requirement-text lookup; a missing/malformed spec.md here is an
// unexpected failure, since Resolve already confirmed the Spec exists.
func readSpecBody(root, specDir string) ([]byte, error) {
	return artifacts.ReadBody(filepath.Join(root, specDir, "spec.md"))
}

// readPlanBodyBestEffort reads the owning Spec's own plan.md body, or
// nil when it doesn't exist yet — a Spec without a Plan still prepares
// successfully, simply with no Plan sections associated (spec Edge
// Cases).
func readPlanBodyBestEffort(root, specDir string) []byte {
	body, err := artifacts.ReadBody(filepath.Join(root, specDir, "plan.md"))
	if err != nil {
		return nil
	}
	return body
}

// parseTaskNumberArg parses --task's own value, which may be a bare
// "TASK-###" or a composite "SPEC-###:TASK-###" — a composite naming a
// different Spec than the positional argument is rejected as
// invalid_argument (contracts §1).
func parseTaskNumberArg(raw string, spec ids.EntityID, cfg project.Configuration) (int, error) {
	if idx := strings.Index(raw, ":"); idx >= 0 {
		specPart, taskPart := raw[:idx], raw[idx+1:]
		specID, err := ids.Parse(ids.Spec, specPart, cfg.IDWidth)
		if err != nil {
			return 0, fmt.Errorf("%w: %q's Spec half: %v", ErrInvalidArgument, raw, err)
		}
		if specID != spec {
			return 0, fmt.Errorf("%w: --task %q names a different Spec than %s", ErrInvalidArgument, raw, spec)
		}
		raw = taskPart
	}
	id, err := ids.Parse(ids.Task, raw, cfg.IDWidth)
	if err != nil {
		return 0, fmt.Errorf("%w: --task %q: %v", ErrInvalidArgument, raw, err)
	}
	return id.Number, nil
}

// renderPreparation shapes one TaskPreparation as JSON-ready data per
// contracts §1.
func renderPreparation(p prepare.TaskPreparation, spec ids.EntityID) map[string]any {
	requirements := make([]map[string]any, 0, len(p.Requirements))
	for _, r := range p.Requirements {
		requirements = append(requirements, map[string]any{
			"ref":         r.Ref.String(),
			"content":     r.Content,
			"fingerprint": r.Fingerprint,
		})
	}

	planSections := make([]map[string]any, 0, len(p.PlanSections))
	for _, s := range p.PlanSections {
		matched := make([]string, 0, len(s.MatchedRequirements))
		for _, m := range s.MatchedRequirements {
			matched = append(matched, m.String())
		}
		planSections = append(planSections, map[string]any{
			"heading":              s.Heading,
			"content":              s.Content,
			"fingerprint":          s.Fingerprint,
			"matched_requirements": matched,
		})
	}

	blockers := make([]string, 0, len(p.Readiness.Blockers))
	for _, b := range p.Readiness.Blockers {
		blockers = append(blockers, fmt.Sprintf("%s:%s", spec, b))
	}

	// code_context reuses renderPackageItemJSON — the same per-item
	// shape --mode package already renders (033) — so a caller reading
	// internal prepare's own JSON learns one item shape, not two
	// (044-architecture-code-context-rules contracts §6).
	codeContext := make([]map[string]any, 0, len(p.CodeContext))
	for _, item := range p.CodeContext {
		codeContext = append(codeContext, renderPackageItemJSON(item, false, false))
	}
	codeScopeNotFound := p.CodeScopeNotFound
	if codeScopeNotFound == nil {
		codeScopeNotFound = []string{}
	}

	return map[string]any{
		"task":                 fmt.Sprintf("%s:%s", spec, p.Task),
		"heading":              p.Heading,
		"ready":                p.Readiness.Ready,
		"blockers":             blockers,
		"requirements":         requirements,
		"scope":                p.Scope,
		"verify":               p.Verify,
		"plan_sections":        planSections,
		"code_context":         codeContext,
		"code_scope_not_found": codeScopeNotFound,
	}
}
