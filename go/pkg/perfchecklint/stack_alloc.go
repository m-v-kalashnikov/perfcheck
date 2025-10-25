package perfchecklint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

const smallStackThreshold = 32

var stackAllocAnalyzer = &analysis.Analyzer{
	Name:     "perf_prefer_stack_alloc",
	Doc:      "reports heap allocations of tiny structs/values",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_prefer_stack_alloc")
		if !ok {
			return nil, fmt.Errorf("rule perf_prefer_stack_alloc not found")
		}

		if pass.TypesSizes == nil {
			return nil, fmt.Errorf("type sizes unavailable")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		nodeFilter := []ast.Node{(*ast.UnaryExpr)(nil), (*ast.CallExpr)(nil)}
		ins.Preorder(nodeFilter, func(node ast.Node) {
			switch v := node.(type) {
			case *ast.UnaryExpr:
				if v.Op != token.AND {
					return
				}
				reportSmallComposite(pass, v, v.X, rule)
			case *ast.CallExpr:
				if ident, ok := v.Fun.(*ast.Ident); !ok || ident.Name != "new" {
					return
				}
				reportSmallNew(pass, v, rule)
			}
		})

		return nil, nil
	},
}

func reportSmallComposite(pass *analysis.Pass, unary *ast.UnaryExpr, expr ast.Expr, rule ruleset.Rule) {
	typ := pass.TypesInfo.TypeOf(expr)
	if typ == nil {
		return
	}
	if !isSmallStackCandidate(typ) {
		return
	}
	size := pass.TypesSizes.Sizeof(typ)
	if size <= 0 || size > smallStackThreshold {
		return
	}
	msg := fmt.Sprintf("%s is %dB; prefer stack allocation", types.TypeString(typ, nil), size)
	report(pass, unary.OpPos, rule, msg)
}

func reportSmallNew(pass *analysis.Pass, call *ast.CallExpr, rule ruleset.Rule) {
	if len(call.Args) != 1 {
		return
	}
	typ := pass.TypesInfo.TypeOf(call)
	ptr, ok := typ.(*types.Pointer)
	if !ok {
		return
	}
	if !isSmallStackCandidate(ptr.Elem()) {
		return
	}
	size := pass.TypesSizes.Sizeof(ptr.Elem())
	if size <= 0 || size > smallStackThreshold {
		return
	}
	msg := fmt.Sprintf("new(%s) allocates %dB on heap; store it by value", types.TypeString(ptr.Elem(), nil), size)
	report(pass, call.Lparen, rule, msg)
}

func isSmallStackCandidate(typ types.Type) bool {
	if typ == nil {
		return false
	}
	under := typ.Underlying()
	switch under.(type) {
	case *types.Struct, *types.Array:
		return true
	case *types.Basic:
		return true
	}
	return false
}
