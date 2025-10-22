package analyzer

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

var reflectionLoopAnalyzer = &analysis.Analyzer{
	Name:     "perf_reflection_dynamic_loop",
	Doc:      "reports reflection usage inside hot loops",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_avoid_reflection_dynamic")
		if !ok {
			return nil, fmt.Errorf("rule perf_avoid_reflection_dynamic not found")
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
			checkReflectionBody(pass, body, rule)
		})

		return nil, nil
	},
}

func checkReflectionBody(pass *analysis.Pass, body *ast.BlockStmt, rule ruleset.Rule) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.CallExpr:
			if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
				if obj := pass.TypesInfo.Uses[sel.Sel]; obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "reflect" {
					report(pass, expr.Pos(), rule, "reflection call inside loop")
					return false
				}
			}
		case *ast.TypeAssertExpr:
			if _, ok := expr.Type.(*ast.InterfaceType); ok {
				return true
			}
			if isGoASTType(expr.Type) {
				return true
			}
			report(pass, expr.Pos(), rule, "type assertion inside loop triggers dynamic dispatch")
			return false
		}
		return true
	})
}

func isGoASTType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return isGoASTType(t.X)
	case *ast.SelectorExpr:
		if pkgIdent, ok := t.X.(*ast.Ident); ok && pkgIdent.Name == "ast" {
			return true
		}
	case *ast.Ident:
		return t.Name == "ast"
	}
	return false
}
