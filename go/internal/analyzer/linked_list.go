package analyzer

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

var linkedListAnalyzer = &analysis.Analyzer{
	Name:     "perf_avoid_linked_list",
	Doc:      "reports container/list usage",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_avoid_linked_list")
		if !ok {
			return nil, fmt.Errorf("rule perf_avoid_linked_list not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		reported := make(map[string]bool)
		ins.Preorder([]ast.Node{(*ast.SelectorExpr)(nil)}, func(node ast.Node) {
			sel, ok := node.(*ast.SelectorExpr)
			if !ok {
				return
			}
			if !isContainerListSelector(pass, sel) {
				return
			}
			pos := pass.Fset.Position(sel.Pos())
			if reported[pos.Filename] {
				return
			}
			reported[pos.Filename] = true
			report(pass, sel.Sel.Pos(), rule, "linked list usage via container/list")
		})

		return nil, nil
	},
}

func isContainerListSelector(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	if sel == nil || sel.Sel == nil || pass.TypesInfo == nil {
		return false
	}
	if obj := pass.TypesInfo.Uses[sel.Sel]; obj != nil {
		if pkg := obj.Pkg(); pkg != nil && pkg.Path() == "container/list" {
			return true
		}
	}
	if selInfo := pass.TypesInfo.Selections[sel]; selInfo != nil {
		if pkg := selInfo.Obj().Pkg(); pkg != nil && pkg.Path() == "container/list" {
			return true
		}
	}
	return false
}
