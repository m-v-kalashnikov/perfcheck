package analyzer

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

var boundConcurrencyAnalyzer = &analysis.Analyzer{
	Name:     "perf_bound_concurrency",
	Doc:      "reports unbounded goroutine creation inside loops",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_bound_concurrency")
		if !ok {
			return nil, fmt.Errorf("rule perf_bound_concurrency not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		loopNodes := []ast.Node{(*ast.ForStmt)(nil), (*ast.RangeStmt)(nil)}

		ins.Preorder(loopNodes, func(node ast.Node) {
			var body *ast.BlockStmt
			switch stmt := node.(type) {
			case *ast.ForStmt:
				body = stmt.Body
			case *ast.RangeStmt:
				body = stmt.Body
			default:
				return
			}
			if body == nil {
				return
			}
			if hasUnboundedGoroutine(body) {
				report(pass, body.Pos(), rule, "goroutine spawned inside loop without bounds")
			}
		})

		return nil, nil
	},
}

func hasUnboundedGoroutine(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.GoStmt); ok {
			// Stop once the first goroutine is found.
			found = true
			if goStmt.Call != nil && goStmt.Call.Fun != nil {
				return false
			}
			return false
		}
		return !found
	})
	return found
}
