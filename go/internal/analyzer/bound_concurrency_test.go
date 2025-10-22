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

func TestBoundConcurrencyAnalyzerFlagsGoInLoop(t *testing.T) {
	src := `package sample

func run(tasks []func()) {
	for _, task := range tasks {
		go task()
	}
}
`
	diags := runBoundConcurrency(src)
	require.Len(t, diags, 1)
	require.True(t, containsRule(diags, "perf_bound_concurrency"))
}

func TestBoundConcurrencyAnalyzerIgnoresOutsideLoop(t *testing.T) {
	src := `package sample

func run(task func()) {
	go task()
}
`
	diags := runBoundConcurrency(src)
	require.Empty(t, diags)
}

func runBoundConcurrency(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "go_spawn.go", src, parser.ParseComments)
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

	rule, ok := ruleset.MustDefault().RuleByID("perf_bound_concurrency")
	if !ok {
		panic("rule not found")
	}

	var diags []analysis.Diagnostic
	pass := analysis.Pass{
		Fset:      fset,
		TypesInfo: info,
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.ForStmt:
			if stmt.Body != nil && hasUnboundedGoroutine(stmt.Body) {
				report(&pass, stmt.Body.Pos(), rule, "goroutine spawned inside loop without bounds")
			}
		case *ast.RangeStmt:
			if stmt.Body != nil && hasUnboundedGoroutine(stmt.Body) {
				report(&pass, stmt.Body.Pos(), rule, "goroutine spawned inside loop without bounds")
			}
		}
		return true
	})
	return diags
}
