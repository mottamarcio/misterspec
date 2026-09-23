package gosource

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// Declaration is one top-level declaration's own parsed shape
// (044-architecture-code-context-rules data-model.md
// "CodeDeclaration", contracts §1).
type Declaration struct {
	Name      string
	Kind      string // "func" | "type" | "const" | "var"
	Signature string // doc comment + signature/declaration text, never the func body
	Body      string // full declaration text, including any func body
	StartLine int
	EndLine   int
}

// Declarations returns every top-level func/type/const/var
// declaration in path, via go/parser's ParseComments mode — signature
// text, doc comment, and file-absolute line range for each
// (research.md #1/#2, contracts §1). For type/const/var declarations
// (which have no separate "body" the way a func does), Signature and
// Body are identical — the signature/body tier split (research.md #7)
// only meaningfully applies to a func's own body, which can be
// arbitrarily large; a short const/var/type declaration has nothing
// smaller to offer.
func Declarations(path string) ([]Declaration, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	decls := make([]Declaration, 0, len(file.Decls))
	for _, d := range file.Decls {
		switch decl := d.(type) {
		case *ast.FuncDecl:
			decls = append(decls, funcDeclaration(fset, src, decl))
		case *ast.GenDecl:
			decls = append(decls, genDeclarations(fset, src, decl)...)
		}
	}
	return decls, nil
}

func funcDeclaration(fset *token.FileSet, src []byte, decl *ast.FuncDecl) Declaration {
	doc := ""
	if decl.Doc != nil {
		doc = decl.Doc.Text()
	}

	sigStart := fset.Position(decl.Pos()).Offset
	sigEnd := len(src)
	if decl.Body != nil {
		sigEnd = fset.Position(decl.Body.Pos()).Offset
	} else {
		sigEnd = fset.Position(decl.End()).Offset
	}
	signatureText := strings.TrimRight(string(src[sigStart:sigEnd]), " \t\n")

	signature := signatureText
	if doc != "" {
		signature = strings.TrimRight(doc, "\n") + "\n" + signatureText
	}

	bodyStart := fset.Position(decl.Pos()).Offset
	bodyEnd := fset.Position(decl.End()).Offset
	body := string(src[bodyStart:bodyEnd])

	return Declaration{
		Name:      decl.Name.Name,
		Kind:      "func",
		Signature: signature,
		Body:      body,
		StartLine: fset.Position(decl.Pos()).Line,
		EndLine:   fset.Position(decl.End()).Line,
	}
}

// genDeclarations expands one GenDecl (which may declare more than one
// identifier, e.g. "const A, B = 1, 2" or a parenthesized "var (...)"
// block) into one Declaration per named identifier.
func genDeclarations(fset *token.FileSet, src []byte, decl *ast.GenDecl) []Declaration {
	kind := ""
	switch decl.Tok {
	case token.TYPE:
		kind = "type"
	case token.CONST:
		kind = "const"
	case token.VAR:
		kind = "var"
	default:
		return nil
	}

	declDoc := ""
	if decl.Doc != nil {
		declDoc = decl.Doc.Text()
	}

	var out []Declaration
	for _, spec := range decl.Specs {
		names := specNames(spec)
		if len(names) == 0 {
			continue
		}

		specStart := fset.Position(spec.Pos()).Offset
		specEnd := fset.Position(spec.End()).Offset
		specText := string(src[specStart:specEnd])

		doc := specDoc(spec)
		if doc == "" && len(decl.Specs) == 1 {
			doc = declDoc
		}
		full := specText
		if doc != "" {
			full = strings.TrimRight(doc, "\n") + "\n" + specText
		}

		startLine := fset.Position(spec.Pos()).Line
		endLine := fset.Position(spec.End()).Line

		for _, name := range names {
			out = append(out, Declaration{
				Name:      name,
				Kind:      kind,
				Signature: full,
				Body:      full,
				StartLine: startLine,
				EndLine:   endLine,
			})
		}
	}
	return out
}

// specDoc returns spec's own doc comment when it has one of its own —
// a parenthesized block's individual specs each carry their own Doc
// separately from the GenDecl's own leading Doc.
func specDoc(spec ast.Spec) string {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if s.Doc != nil {
			return s.Doc.Text()
		}
	case *ast.ValueSpec:
		if s.Doc != nil {
			return s.Doc.Text()
		}
	}
	return ""
}

func specNames(spec ast.Spec) []string {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		return []string{s.Name.Name}
	case *ast.ValueSpec:
		names := make([]string, 0, len(s.Names))
		for _, n := range s.Names {
			if n.Name != "_" {
				names = append(names, n.Name)
			}
		}
		return names
	default:
		return nil
	}
}
