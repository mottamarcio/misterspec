package prepare

import (
	"regexp"
	"strconv"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	contextengine "github.com/mottamarcio/misterspec/internal/context"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/validation"
)

// RequirementText is one served Requirement's own text, for a Task's
// assembled context (034/data-model.md "RequirementText").
type RequirementText struct {
	Ref         validation.RequirementRef
	Content     string
	Fingerprint string
}

// requirementHeadingPattern mirrors internal/validation's own
// (unexported) pattern for a Requirement heading — duplicated rather
// than imported since it is unexported there; both match the same
// kit/templates/spec.md.tmpl convention ("### R1 — <Requirement>").
var requirementHeadingPattern = regexp.MustCompile(`^R(\d+)\b`)

// LookupRequirementTexts returns, for each ref in refs, its own
// Requirement section content from specBody (a Spec's own Markdown
// body), plus a content fingerprint (034/data-model.md
// "RequirementText"). A ref whose number has no matching section in
// specBody contributes nothing — internal/validation's own
// uncovered/unknown-reference checks are the place that flags that as
// a structural problem; prepare simply omits what it cannot find.
func LookupRequirementTexts(specBody []byte, refs []validation.RequirementRef) []RequirementText {
	doc := artifacts.ParseDocument(specBody)

	byNumber := map[int]string{}
	for _, section := range doc.Sections {
		sub := requirementHeadingPattern.FindStringSubmatch(section.Heading)
		if sub == nil {
			continue
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			continue
		}
		byNumber[n] = section.Body
	}

	out := make([]RequirementText, 0, len(refs))
	for _, ref := range refs {
		content, ok := byNumber[ref.Number]
		if !ok {
			continue
		}
		out = append(out, RequirementText{
			Ref:         ref,
			Content:     content,
			Fingerprint: contextengine.Fingerprint(content),
		})
	}
	return out
}

// PlanSectionAssociation is one Plan section whose own content mentions
// a Requirement the Task serves (034/data-model.md
// "PlanSectionAssociation", research.md Decision 4).
type PlanSectionAssociation struct {
	Heading             string
	Content             string
	Fingerprint         string
	MatchedRequirements []validation.RequirementRef
}

// AssociatePlanSections returns every section of planBody (a Spec's own
// plan.md body) whose own content mentions at least one of served's
// Requirement references, in either the bare "R<N>" form or the
// composite "SPEC-###:R#" form — no new Plan-authoring convention is
// required (research.md Decision 4). A section matching none of served
// is not returned at all, never returned with an empty
// MatchedRequirements.
func AssociatePlanSections(planBody []byte, served []validation.RequirementRef) []PlanSectionAssociation {
	doc := artifacts.ParseDocument(planBody)

	var out []PlanSectionAssociation
	for _, section := range doc.Sections {
		if section.Body == "" {
			continue
		}
		var matched []validation.RequirementRef
		for _, ref := range served {
			bare := regexp.MustCompile(`\bR` + strconv.Itoa(ref.Number) + `\b`)
			if bare.MatchString(section.Body) || regexpContains(section.Body, ref.String()) {
				matched = append(matched, ref)
			}
		}
		if len(matched) == 0 {
			continue
		}
		out = append(out, PlanSectionAssociation{
			Heading:             section.Heading,
			Content:             section.Body,
			Fingerprint:         contextengine.Fingerprint(section.Body),
			MatchedRequirements: matched,
		})
	}
	return out
}

// regexpContains reports whether needle appears verbatim in haystack —
// a tiny helper so the composite-form check in AssociatePlanSections
// reads as plainly as the bare-form regex check next to it.
func regexpContains(haystack, needle string) bool {
	return regexp.MustCompile(regexp.QuoteMeta(needle)).MatchString(haystack)
}

// TaskPreparation is the single response bundle prepare returns
// (034/data-model.md "TaskPreparation", spec FR-001).
type TaskPreparation struct {
	Task         ids.EntityID
	Heading      string
	Readiness    TaskReadiness
	Requirements []RequirementText
	Scope        string
	Verify       string
	PlanSections []PlanSectionAssociation
}

// BuildTaskPreparation assembles the full TaskPreparation bundle from
// its already-computed parts (034/data-model.md "TaskPreparation").
// Called once every other piece (readiness, requirement text, fields,
// Plan sections) is available — it performs no I/O and no lookup of
// its own.
func BuildTaskPreparation(task ids.EntityID, heading string, readiness TaskReadiness, requirements []RequirementText, fields TaskFields, planSections []PlanSectionAssociation) TaskPreparation {
	return TaskPreparation{
		Task:         task,
		Heading:      heading,
		Readiness:    readiness,
		Requirements: requirements,
		Scope:        fields.Scope,
		Verify:       fields.Verify,
		PlanSections: planSections,
	}
}
