package analyzer

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

var deferInLoopAnalyzer = &analysis.Analyzer{
	Name:     "perf_no_defer_in_loop",
	Doc:      "reports defer statements inside loops",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_no_defer_in_loop")
		if !ok {
			return nil, fmt.Errorf("rule perf_no_defer_in_loop not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		loopNodes := collectLoopNodes(ins)
		nodeFilter := []ast.Node{(*ast.DeferStmt)(nil)}
		ins.WithStack(nodeFilter, func(node ast.Node, push bool, stack []ast.Node) bool {
			if !push {
				return true
			}
			if !deferInsideLoop(stack, loopNodes) {
				return true
			}
			report(pass, node.Pos(), rule, "defer inside loop delays cleanup until function exit")
			return true
		})

		return nil, nil
	},
}

func collectLoopNodes(ins *inspector.Inspector) map[ast.Node]struct{} {
	loops := make(map[ast.Node]struct{})
	loopFilter := []ast.Node{(*ast.ForStmt)(nil), (*ast.RangeStmt)(nil)}
	ins.Preorder(loopFilter, func(node ast.Node) {
		loops[node] = struct{}{}
	})
	return loops
}

func deferInsideLoop(stack []ast.Node, loops map[ast.Node]struct{}) bool {
	for _, ancestor := range stack {
		if _, ok := loops[ancestor]; ok {
			return true
		}
	}
	return false
}
