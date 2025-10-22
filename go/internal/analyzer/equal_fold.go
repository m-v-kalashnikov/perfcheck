package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var equalFoldAnalyzer = &analysis.Analyzer{
	Name:     "perf_equal_fold_compare",
	Doc:      "reports case-insensitive comparisons built via ToLower/ToUpper",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (interface{}, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_equal_fold_compare")
		if !ok {
			return nil, fmt.Errorf("rule perf_equal_fold_compare not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		ins.Preorder([]ast.Node{(*ast.BinaryExpr)(nil)}, func(node ast.Node) {
			bin := node.(*ast.BinaryExpr)
			if bin.Op != token.EQL && bin.Op != token.NEQ {
				return
			}
			if isStringsNormalizeCall(pass, bin.X) || isStringsNormalizeCall(pass, bin.Y) {
				report(pass, bin.Pos(), rule, "use strings.EqualFold for case-insensitive comparison")
			}
		})

		return nil, nil
	},
}

func isStringsNormalizeCall(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	if obj.Pkg().Path() != "strings" {
		return false
	}
	name := obj.Name()
	return (name == "ToLower" || name == "ToUpper") && len(call.Args) == 1
}
