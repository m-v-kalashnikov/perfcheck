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

var syncPoolPointerAnalyzer = &analysis.Analyzer{
	Name:     "perf_syncpool_store_pointers",
	Doc:      "reports storing non-pointer values in sync.Pool",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_syncpool_store_pointers")
		if !ok {
			return nil, fmt.Errorf("rule perf_syncpool_store_pointers not found")
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
			if len(call.Args) != 1 {
				return
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil || sel.Sel.Name != "Put" {
				return
			}
			obj := pass.TypesInfo.Uses[sel.Sel]
			fn, ok := obj.(*types.Func)
			if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "sync" {
				return
			}

			argType := pass.TypesInfo.TypeOf(call.Args[0])
			if argType == nil {
				return
			}
			if isPointerLike(argType) {
				return
			}
			report(
				pass,
				call.Args[0].Pos(),
				rule,
				"store pointer types in sync.Pool to avoid interface allocations",
			)
		})

		return nil, nil
	},
}

func isPointerLike(t types.Type) bool {
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Interface, *types.Slice, *types.Map, *types.Chan, *types.Signature:
		return true
	}
	// Accept untyped nil as pointer-like.
	if basic, ok := t.(*types.Basic); ok && basic.Kind() == types.UntypedNil {
		return true
	}
	return false
}
