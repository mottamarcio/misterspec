// Command gen regenerates every canonical Skill's kit/skills/<name>/SKILL.md
// from internal/skillgen's Manifests + KnownFragments. It is a
// development-time authoring tool, invoked via
// `go generate ./internal/skillgen/...` (see the //go:generate
// directive in internal/skillgen/generate.go) — never wired into the
// product's own CLI (plan.md "Scale/Scope").
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mottamarcio/misterspec/internal/skillgen"
)

// repoRoot locates the repository root relative to this source file
// (internal/skillgen/cmd/gen/main.go is four directories below root),
// so the tool writes correct output regardless of the caller's own
// working directory.
func repoRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("gen: could not determine source file location")
	}
	// this file: <root>/internal/skillgen/cmd/gen/main.go
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")), nil
}

func main() {
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, m := range skillgen.Manifests {
		out, err := skillgen.Generate(m, skillgen.KnownFragments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen: %v\n", err)
			os.Exit(1)
		}
		path := filepath.Join(root, "kit", "skills", m.SkillName, "SKILL.md")
		if err := os.WriteFile(path, out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "gen: writing %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Println("wrote", path)
	}
}
