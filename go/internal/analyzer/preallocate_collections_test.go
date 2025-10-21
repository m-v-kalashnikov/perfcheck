package analyzer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yourname/perfcheck/go/internal/ruleset"
	"golang.org/x/tools/go/analysis"
)

func TestPreallocateCollectionsDetectsMissingReserve(t *testing.T) {
	src := `package sample

func build(items []int) []int {
	var out []int
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
`
	diags := runPreallocate(src)
	require.Len(t, diags, 1)
	require.True(t, containsRule(diags, "perf_preallocate_collections"))
}

func TestPreallocateCollectionsIgnoresReservedSlice(t *testing.T) {
	src := `package sample

func build(items []int) []int {
	out := make([]int, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
`
	diags := runPreallocate(src)
	require.Empty(t, diags)
}

func TestPreallocateCollectionsIgnoresSizedMake(t *testing.T) {
	src := `package sample

func build(items []int) []int {
	out := make([]int, len(items))
	for i, item := range items {
		out[i] = item
	}
	return out
}
`
	diags := runPreallocate(src)
	require.Empty(t, diags)
}

func runPreallocate(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "preallocate.go", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	conf := types.Config{Importer: importer.Default()}
	if _, err := conf.Check("sample", fset, []*ast.File{file}, info); err != nil {
		panic(err)
	}

	var diags []analysis.Diagnostic
	rule, ok := ruleset.MustDefault().RuleByID("perf_preallocate_collections")
	if !ok {
		panic("rule not found")
	}
	var fnBody *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			fnBody = fn.Body
			break
		}
	}
	if fnBody == nil {
		panic("no function body found")
	}

	pass := analysis.Pass{
		Fset:      fset,
		TypesInfo: info,
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
	}
	analyzePreallocation(&pass, fnBody, rule)
	return diags
}
