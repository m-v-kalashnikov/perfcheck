package analyzer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yourname/perfcheck/go/internal/ruleset"
	"golang.org/x/tools/go/analysis"
)

func TestStringConcatLoopAnalyzerFindsViolation(t *testing.T) {
	src := `package sample

func build(items []string) string {
	s := ""
	for _, item := range items {
		s += item
	}
	return s
}
`

	diags := runConcat(src)
	require.Len(t, diags, 1)
	require.True(t, containsRule(diags, "perf_avoid_string_concat_loop"))
}

func TestStringConcatLoopAnalyzerIgnoresBuilder(t *testing.T) {
	src := `package sample

import "strings"

func build(items []string) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(item)
	}
	return b.String()
}
`

	diags := runConcat(src)
	require.Empty(t, diags)
}

func runConcat(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "concat.go", src, parser.ParseComments)
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
	rule, ok := ruleset.MustDefault().RuleByID("perf_avoid_string_concat_loop")
	if !ok {
		panic("rule not found")
	}
	pass := analysis.Pass{
		Fset:      fset,
		TypesInfo: info,
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
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
	checkConcatBody(&pass, fnBody, rule)
	return diags
}

func containsRule(diags []analysis.Diagnostic, id string) bool {
	for _, d := range diags {
		if strings.Contains(d.Message, id) {
			return true
		}
	}
	return false
}
