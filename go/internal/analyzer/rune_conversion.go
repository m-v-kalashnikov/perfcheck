package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

var runeConversionAnalyzer = &analysis.Analyzer{
	Name:     "perf_avoid_rune_conversion",
	Doc:      "reports []rune conversions used only for ranging",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_avoid_rune_conversion")
		if !ok {
			return nil, fmt.Errorf("rule perf_avoid_rune_conversion not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		nodeFilter := []ast.Node{(*ast.RangeStmt)(nil)}
		ins.Preorder(nodeFilter, func(node ast.Node) {
			rangeStmt, ok := node.(*ast.RangeStmt)
			if !ok {
				return
			}
			call, ok := rangeStmt.X.(*ast.CallExpr)
			if !ok || !isRuneSliceConversion(pass, call) {
				return
			}
			report(
				pass,
				call.Lparen,
				rule,
				"convert string to []rune only once; iterate the string directly",
			)
		})

		return nil, nil
	},
}

func isRuneSliceConversion(pass *analysis.Pass, call *ast.CallExpr) bool {
	if call == nil || pass.TypesInfo == nil {
		return false
	}
	arrayType, ok := call.Fun.(*ast.ArrayType)
	if !ok {
		return false
	}
	if arrayType.Len != nil {
		return false
	}
	elt, ok := arrayType.Elt.(*ast.Ident)
	if !ok || elt.Name != "rune" {
		return false
	}
	if len(call.Args) != 1 {
		return false
	}
	typ := pass.TypesInfo.TypeOf(call.Args[0])
	if typ == nil {
		return false
	}
	return types.Identical(typ, types.Typ[types.String])
}
