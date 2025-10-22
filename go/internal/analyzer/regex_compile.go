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

var regexCompileLoopAnalyzer = &analysis.Analyzer{
	Name:     "perf_regex_compile_loop",
	Doc:      "reports regexp compilation executed inside loops",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_regex_compile_once")
		if !ok {
			return nil, fmt.Errorf("rule perf_regex_compile_once not found")
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
			checkRegexCalls(pass, body, rule)
		})

		return nil, nil
	},
}

func checkRegexCalls(pass *analysis.Pass, body *ast.BlockStmt, rule ruleset.Rule) {
	targets := map[string]struct{}{
		"Compile":          {},
		"MustCompile":      {},
		"CompilePOSIX":     {},
		"MustCompilePOSIX": {},
	}

	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		obj := pass.TypesInfo.Uses[sel.Sel]
		fn, ok := obj.(*types.Func)
		if !ok || fn.Pkg() == nil {
			return true
		}
		if fn.Pkg().Path() != "regexp" {
			return true
		}
		if _, match := targets[fn.Name()]; !match {
			return true
		}

		pass.Report(analysis.Diagnostic{
			Pos:      call.Pos(),
			Message:  fmt.Sprintf("[%s] compile regexp outside loops to avoid repeated parsing", rule.ID),
			Category: rule.Category,
		})
		return true
	})
}
