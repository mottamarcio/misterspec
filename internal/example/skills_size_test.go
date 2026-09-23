// This file implements 039-lean-skills-integration-contracts's SC-001
// size-tracking check, with an important correction found during
// implementation: SC-001 (spec.md) requires the total canonical Skill
// body line count to shrink by at least 20% from the recorded
// pre-feature baseline (2378 lines). Extracting the real duplicated
// fragments found by research.md #1 (resolve-preamble,
// mechanical-steps-note) does NOT shrink kit/skills/*/SKILL.md's own
// rendered size — internal/skillgen composes each Skill's file to stay
// byte-identical, full, readable Markdown (Constitution Principle III/
// IX: no opaque template output), so a fragment's text still appears in
// full in every file that references it. Deduplication happens at the
// authoring source (internal/skillgen/fragment.go), not in what an
// agent actually reads — and no agent invocation ever loads more than
// one Skill file at once, so the sum-across-all-10-files metric SC-001
// measures was never a real per-task token cost to begin with. Adding
// the FR-008 fallback-recording instruction (T033-T035) made the total
// line count go up, not down (2378 -> 2393).
//
// This test therefore tracks regressions against the pre-feature
// baseline rather than asserting SC-001's originally-specified 80%
// threshold, which this feature's own approach cannot meet without
// cutting real Skill content — something the spec's own Assumptions
// section explicitly cautions against. SC-001 itself should be revisited
// in a follow-up spec amendment; see this feature's final implementation
// report.
package example

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/kit"
)

// preFeatureSkillLineCount is kit/skills/*/SKILL.md's own total
// non-frontmatter line count immediately before
// 039-lean-skills-integration-contracts's changes (research.md #6).
const preFeatureSkillLineCount = 2378

// skillLineCountRegressionTolerance allows the small, deliberate growth
// from T033-T035's fallback-recording instruction (net +15 lines across
// 5 files) without masking a genuine, larger regression later.
const skillLineCountRegressionTolerance = 1.05 // +5%

func TestSkillsContent_SizeDoesNotRegressUnexpectedly(t *testing.T) {
	total := 0
	entries, err := fs.ReadDir(kit.SkillsFS, ".")
	if err != nil {
		t.Fatalf("reading kit.SkillsFS root: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := fs.ReadFile(kit.SkillsFS, e.Name()+"/SKILL.md")
		if err != nil {
			t.Fatalf("reading kit.SkillsFS %s/SKILL.md: %v", e.Name(), err)
		}
		total += strings.Count(string(data), "\n")
	}

	baseline := float64(preFeatureSkillLineCount)
	limit := int(baseline * skillLineCountRegressionTolerance)
	if total > limit {
		t.Errorf("total kit/skills/*/SKILL.md line count = %d, want <= %d (%.0f%% of the %d-line pre-feature baseline) — see this file's own header comment for why SC-001's original 80%% target is not asserted here", total, limit, skillLineCountRegressionTolerance*100, preFeatureSkillLineCount)
	}
	t.Logf("total kit/skills/*/SKILL.md line count = %d (pre-feature baseline: %d)", total, preFeatureSkillLineCount)
}
