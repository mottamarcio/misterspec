package architecture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/mottamarcio/misterspec/internal/gosource"
	"github.com/mottamarcio/misterspec/internal/project"
)

// CheckArchitecture evaluates rules against root's own source tree.
// When root has no go.mod, every rule in rules produces exactly one
// Result with Status "not_evaluated" and Reason
// "no_adapter_for_project" (research.md #8, spec FR-004). exclusions
// are project.Configuration's own CodeExclusions glob patterns
// (research.md #4) — a matching path is never walked, regardless of
// whether it would otherwise violate a rule.
func CheckArchitecture(root string, rules []project.ArchitectureRule, exclusions []string) (Report, error) {
	modulePath, hasAdapter, err := detectGoModule(root)
	if err != nil {
		return Report{}, err
	}
	if !hasAdapter {
		return Report{Results: notEvaluatedAll(rules, ReasonNoAdapterForProject)}, nil
	}

	files, err := gosource.WalkGoFiles(root, exclusions)
	if err != nil {
		return Report{}, err
	}

	results := make([]Result, 0, len(rules))
	for i, rule := range rules {
		switch rule.Kind {
		case "forbidden_dependency", "layer_boundary":
			fails, err := evaluateDependencyRule(root, modulePath, files, rule)
			if err != nil {
				return Report{}, err
			}
			if len(fails) == 0 {
				results = append(results, Result{RuleIndex: i, Status: StatusPass})
				continue
			}
			for _, f := range fails {
				f.RuleIndex = i
				results = append(results, f)
			}
		default:
			results = append(results, Result{
				RuleIndex: i,
				Status:    StatusNotEvaluated,
				Reason:    ReasonRuleKindUnsupported,
				Message:   fmt.Sprintf("rule kind %q is not supported by the go adapter", rule.Kind),
			})
		}
	}

	return Report{Adapter: "go", Results: results}, nil
}

// detectGoModule reports root's own module path and whether a go.mod
// exists at all (research.md #8) — the sole applicability signal for
// this feature's one adapter.
func detectGoModule(root string) (modulePath string, hasAdapter bool, err error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return modfile.ModulePath(data), true, nil
}

// notEvaluatedAll renders every rule as not_evaluated with reason —
// used when the whole project has no applicable adapter.
func notEvaluatedAll(rules []project.ArchitectureRule, reason string) []Result {
	out := make([]Result, 0, len(rules))
	for i := range rules {
		out = append(out, Result{RuleIndex: i, Status: StatusNotEvaluated, Reason: reason})
	}
	return out
}

// evaluateDependencyRule checks every file in files matching rule.From
// for an import matching rule.To (after normalizing the import against
// modulePath — research.md #2), returning one fail Result per
// violating import site, in deterministic (file, then line) order.
func evaluateDependencyRule(root, modulePath string, files []string, rule project.ArchitectureRule) ([]Result, error) {
	var fails []Result
	for _, relPath := range files {
		if !gosource.MatchGlob(rule.From, relPath) {
			continue
		}
		imports, err := gosource.Imports(filepath.Join(root, relPath))
		if err != nil {
			return nil, fmt.Errorf("architecture: parsing imports for %s: %w", relPath, err)
		}
		for _, imp := range imports {
			normalized := normalizeImport(modulePath, imp.Path)
			if !gosource.MatchGlob(rule.To, normalized) {
				continue
			}
			fails = append(fails, Result{
				Status:  StatusFail,
				Path:    relPath,
				Line:    imp.Line,
				Message: fmt.Sprintf("%s imports %s, which this rule forbids", relPath, imp.Path),
			})
		}
	}
	return fails, nil
}

// normalizeImport strips modulePath's own prefix from importPath when
// importPath is internal to this project (e.g.
// "fixture.example/b" -> "b" for modulePath "fixture.example"),
// letting a declared rule's own "to" pattern name an internal path
// without repeating the module path — an external import (no matching
// prefix) is left unchanged.
func normalizeImport(modulePath, importPath string) string {
	if modulePath == "" {
		return importPath
	}
	if importPath == modulePath {
		return ""
	}
	if rel, ok := strings.CutPrefix(importPath, modulePath+"/"); ok {
		return rel
	}
	return importPath
}
