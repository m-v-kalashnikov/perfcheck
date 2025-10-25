package perfchecklint

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

var writerPreferBytesAnalyzer = &analysis.Analyzer{
	Name:     "perf_writer_prefer_bytes",
	Doc:      "reports string conversions when writing byte slices",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_writer_prefer_bytes")
		if !ok {
			return nil, fmt.Errorf("rule perf_writer_prefer_bytes not found")
		}

		insAnalyser, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if !ok || insAnalyser == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		insAnalyser.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil {
				return
			}
			name := sel.Sel.Name
			var arg ast.Expr
			switch name {
			case "WriteString":
				if len(call.Args) == 0 {
					return
				}
				arg = call.Args[len(call.Args)-1]
			case "Write":
				if len(call.Args) == 0 {
					return
				}
				arg = call.Args[0]
			default:
				return
			}

			if !isStringConversion(pass.TypesInfo, arg) {
				return
			}

			if _, ok := pass.TypesInfo.TypeOf(arg).(*types.Basic); !ok {
				return
			}

			report(
				pass,
				arg.Pos(),
				rule,
				"avoid converting []byte to string when writing; use byte-oriented writes",
			)
		})

		return nil, nil
	},
}

func isStringConversion(info *types.Info, expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	if ident, ok := call.Fun.(*ast.Ident); !ok || ident.Name != "string" {
		return false
	}
	typ := info.TypeOf(expr)
	if typ == nil {
		return false
	}
	if basic, ok := typ.(*types.Basic); !ok || basic.Kind() != types.String {
		return false
	}
	argType := info.TypeOf(call.Args[0])
	if argType == nil {
		return false
	}
	if slice, ok := argType.Underlying().(*types.Slice); ok {
		if elem, ok := slice.Elem().(*types.Basic); ok &&
			(elem.Kind() == types.Byte || elem.Kind() == types.Uint8 || elem.Kind() == types.Int32 || elem.Kind() == types.Rune) {
			return true
		}
	}
	return false
}
