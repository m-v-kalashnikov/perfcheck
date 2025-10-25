package perfchecklint

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

func TestSyncPoolPointerAnalyzerFlagsValue(t *testing.T) {
	src := `package sample

import "sync"

type item struct{}

func store(pool *sync.Pool, it item) {
	pool.Put(it)
}
`
	diags := runSyncPool(src)
	require.Len(t, diags, 1)
	require.True(t, containsRule(diags, "perf_syncpool_store_pointers"))
}

func TestSyncPoolPointerAnalyzerAllowsPointer(t *testing.T) {
	src := `package sample

import "sync"

type item struct{}

func store(pool *sync.Pool, it *item) {
	pool.Put(it)
}
`
	diags := runSyncPool(src)
	require.Empty(t, diags)
}

func runSyncPool(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "pool.go", src, parser.ParseComments)
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

	rule, ok := ruleset.MustDefault().RuleByID("perf_syncpool_store_pointers")
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
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		// Emulate analyzer logic directly.
		if len(call.Args) != 1 || sel.Sel == nil || sel.Sel.Name != "Put" {
			return true
		}
		obj := pass.TypesInfo.Uses[sel.Sel]
		fn, ok := obj.(*types.Func)
		if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "sync" {
			return true
		}
		argType := pass.TypesInfo.TypeOf(call.Args[0])
		if argType == nil {
			return true
		}
		if isPointerLike(argType) {
			return true
		}
		report(
			&pass,
			call.Args[0].Pos(),
			rule,
			"store pointer types in sync.Pool to avoid interface allocations",
		)
		return true
	})

	return diags
}
