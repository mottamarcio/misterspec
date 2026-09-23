package gosource

import (
	"go/parser"
	"go/token"
	"strconv"
)

// ImportRef is one import statement's own path and file-absolute line
// (044-architecture-code-context-rules data-model.md "CodeFile"/
// contracts §1).
type ImportRef struct {
	Path string
	Line int
}

// Imports returns path's own import list, parsed via go/parser's
// ImportsOnly mode — cheap, no type-checking (research.md #1/#2).
func Imports(path string) ([]ImportRef, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	refs := make([]ImportRef, 0, len(file.Imports))
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		refs = append(refs, ImportRef{
			Path: p,
			Line: fset.Position(imp.Path.Pos()).Line,
		})
	}
	return refs, nil
}
