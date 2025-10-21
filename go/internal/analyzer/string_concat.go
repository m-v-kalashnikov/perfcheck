package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/yourname/perfcheck/go/internal/ruleset"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var stringConcatLoopAnalyzer = &analysis.Analyzer{
	Name:     "perf_string_concat_loop",
	Doc:      "reports string concatenation inside loops",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (interface{}, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_avoid_string_concat_loop")
		if !ok {
			return nil, fmt.Errorf("rule perf_avoid_string_concat_loop not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		nodeFilter := []ast.Node{(*ast.ForStmt)(nil), (*ast.RangeStmt)(nil)}

		ins.Preorder(nodeFilter, func(node ast.Node) {
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
			checkConcatBody(pass, body, rule)
		})

		return nil, nil
	},
}

func checkConcatBody(pass *analysis.Pass, body *ast.BlockStmt, rule ruleset.Rule) {
	ast.Inspect(body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		switch assign.Tok {
		case token.ADD_ASSIGN:
			if len(assign.Lhs) != 1 {
				return true
			}
			if isString(pass.TypesInfo, assign.Lhs[0]) {
				report(pass, assign.Pos(), rule, "string concatenation using '+=' inside loop")
			}
		case token.ASSIGN:
			if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
				return true
			}
			bin, ok := assign.Rhs[0].(*ast.BinaryExpr)
			if !ok || bin.Op != token.ADD {
				return true
			}
			if !isString(pass.TypesInfo, assign.Lhs[0]) {
				return true
			}
			if exprEqual(assign.Lhs[0], bin.X) {
				report(pass, assign.Pos(), rule, "string concatenation using '=', consider strings.Builder")
			}
		}

		return true
	})
}

func isString(info *types.Info, expr ast.Expr) bool {
	if info == nil {
		return false
	}
	t := info.TypeOf(expr)
	return t != nil && types.Identical(t, types.Typ[types.String])
}

func exprEqual(a, b ast.Expr) bool {
	if a == nil || b == nil {
		return false
	}
	return types.ExprString(a) == types.ExprString(b)
}

func report(pass *analysis.Pass, pos token.Pos, rule ruleset.Rule, msg string) {
	pass.Report(analysis.Diagnostic{
		Pos:      pos,
		Message:  fmt.Sprintf("[%s] %s", rule.ID, msg),
		Category: rule.Category,
	})
}
