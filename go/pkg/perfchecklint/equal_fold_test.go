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

func TestEqualFoldAnalyzerFlagsToLowerCompare(t *testing.T) {
	src := `package sample

import "strings"

func equal(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}
`
	diags := runEqualFold(src)
	require.Len(t, diags, 1)
	require.True(t, containsRule(diags, "perf_equal_fold_compare"))
}

func TestEqualFoldAnalyzerAllowsEqualFold(t *testing.T) {
	src := `package sample

import "strings"

func equal(a, b string) bool {
	return strings.EqualFold(a, b)
}
`
	diags := runEqualFold(src)
	require.Empty(t, diags)
}

func runEqualFold(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "equalfold.go", src, parser.ParseComments)
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

	rule, ok := ruleset.MustDefault().RuleByID("perf_equal_fold_compare")
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
		bin, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		if bin.Op != token.EQL && bin.Op != token.NEQ {
			return true
		}
		if isStringsNormalizeCall(&pass, bin.X) || isStringsNormalizeCall(&pass, bin.Y) {
			report(&pass, bin.Pos(), rule, "use strings.EqualFold for case-insensitive comparison")
		}
		return true
	})

	return diags
}
