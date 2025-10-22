package analyzer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

func TestReflectionLoopAnalyzerFlagsReflectUsage(t *testing.T) {
	src := `package sample

import "reflect"

func check(values []interface{}) {
	for _, v := range values {
		_ = reflect.ValueOf(v).Kind()
	}
}
`

	diags := runReflectionAnalyzer(src)
	require.Len(t, diags, 1)
	require.True(t, containsRule(diags, "perf_avoid_reflection_dynamic"))
}

func TestReflectionLoopAnalyzerAllowsOutsideLoop(t *testing.T) {
	src := `package sample

import "reflect"

func check(v interface{}) reflect.Kind {
	return reflect.ValueOf(v).Kind()
}
`
	diags := runReflectionAnalyzer(src)
	require.Empty(t, diags)
}

func runReflectionAnalyzer(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "reflect.go", src, parser.ParseComments)
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

	rule, ok := ruleset.MustDefault().RuleByID("perf_avoid_reflection_dynamic")
	if !ok {
		panic("rule not found")
	}

	var diags []analysis.Diagnostic
	var fnBody *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			fnBody = fn.Body
			break
		}
	}
	if fnBody == nil {
		panic("no function body")
	}

	pass := analysis.Pass{
		Fset:      fset,
		TypesInfo: info,
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
	}

	ast.Inspect(fnBody, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.ForStmt:
			if stmt.Body != nil {
				checkReflectionBody(&pass, stmt.Body, rule)
			}
		case *ast.RangeStmt:
			if stmt.Body != nil {
				checkReflectionBody(&pass, stmt.Body, rule)
			}
		}
		return true
	})
	return diags
}
