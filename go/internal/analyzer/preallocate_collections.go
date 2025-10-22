package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var preallocateCollectionsAnalyzer = &analysis.Analyzer{
	Name:     "perf_preallocate_collections",
	Doc:      "reports slice growth in range loops without prior preallocation",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (interface{}, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_preallocate_collections")
		if !ok {
			return nil, fmt.Errorf("rule perf_preallocate_collections not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		nodeFilter := []ast.Node{
			(*ast.FuncDecl)(nil),
			(*ast.FuncLit)(nil),
		}

		ins.Preorder(nodeFilter, func(node ast.Node) {
			var body *ast.BlockStmt
			switch fn := node.(type) {
			case *ast.FuncDecl:
				body = fn.Body
			case *ast.FuncLit:
				body = fn.Body
			}
			if body == nil {
				return
			}
			analyzePreallocation(pass, body, rule)
		})

		return nil, nil
	},
}

func analyzePreallocation(pass *analysis.Pass, body *ast.BlockStmt, rule ruleset.Rule) {
	env := make(map[string]bool, 8)
	scanBlock(pass, body, env, false, rule)
}

func scanBlock(pass *analysis.Pass, block *ast.BlockStmt, reserved map[string]bool, insideLoop bool, rule ruleset.Rule) {
	if block == nil {
		return
	}
	for _, stmt := range block.List {
		scanStmt(pass, stmt, reserved, insideLoop, rule)
	}
}

func scanStmt(pass *analysis.Pass, stmt ast.Stmt, reserved map[string]bool, insideLoop bool, rule ruleset.Rule) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if insideLoop {
			checkAppend(pass, s, reserved, rule)
		}
		updateReservedFromAssign(s, reserved)
	case *ast.DeclStmt:
		updateReservedFromDecl(s, reserved)
	case *ast.RangeStmt:
		loopInside := insideLoop
		if !loopInside && isSliceRange(pass.TypesInfo, s.X) {
			loopInside = true
		}
		loopReserved := cloneReserved(reserved)
		// Range statements may declare new variables; process body accordingly.
		scanBlock(pass, s.Body, loopReserved, loopInside, rule)
	case *ast.ForStmt:
		loopReserved := cloneReserved(reserved)
		if s.Init != nil {
			scanStmt(pass, s.Init, loopReserved, insideLoop, rule)
		}
		scanBlock(pass, s.Body, loopReserved, insideLoop, rule)
		if s.Post != nil {
			scanStmt(pass, s.Post, loopReserved, insideLoop, rule)
		}
	case *ast.BlockStmt:
		scanBlock(pass, s, cloneReserved(reserved), insideLoop, rule)
	case *ast.IfStmt:
		if s.Init != nil {
			scanStmt(pass, s.Init, cloneReserved(reserved), insideLoop, rule)
		}
		scanBlock(pass, s.Body, cloneReserved(reserved), insideLoop, rule)
		if elseBlock := asBlock(s.Else); elseBlock != nil {
			scanBlock(pass, elseBlock, cloneReserved(reserved), insideLoop, rule)
		}
	case *ast.SwitchStmt:
		if s.Init != nil {
			scanStmt(pass, s.Init, cloneReserved(reserved), insideLoop, rule)
		}
		for _, stmt := range s.Body.List {
			if clause, ok := stmt.(*ast.CaseClause); ok {
				scanBlock(pass, &ast.BlockStmt{List: clause.Body}, cloneReserved(reserved), insideLoop, rule)
			}
		}
	case *ast.TypeSwitchStmt:
		if s.Init != nil {
			scanStmt(pass, s.Init, cloneReserved(reserved), insideLoop, rule)
		}
		if s.Assign != nil {
			scanStmt(pass, s.Assign, cloneReserved(reserved), insideLoop, rule)
		}
		for _, stmt := range s.Body.List {
			if clause, ok := stmt.(*ast.CaseClause); ok {
				scanBlock(pass, &ast.BlockStmt{List: clause.Body}, cloneReserved(reserved), insideLoop, rule)
			}
		}
	default:
		// no-op
	}
}

func checkAppend(pass *analysis.Pass, assign *ast.AssignStmt, reserved map[string]bool, rule ruleset.Rule) {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	call := extractCall(assign.Rhs[0], "append")
	if call == nil || len(call.Args) == 0 {
		return
	}

	targetKey := exprKey(call.Args[0])
	if targetKey == "" {
		return
	}

	lhsKey := exprKey(assign.Lhs[0])
	if lhsKey == "" || lhsKey != targetKey {
		return
	}

	if reserved[targetKey] {
		return
	}

	if !isSliceType(pass.TypesInfo, assign.Lhs[0]) {
		return
	}

	report(pass, assign.Pos(), rule, "append inside loop without preallocated capacity")
}

func updateReservedFromAssign(assign *ast.AssignStmt, reserved map[string]bool) {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	call := extractCall(assign.Rhs[0], "make")
	if call == nil {
		return
	}

	if !isPreallocMake(call) {
		return
	}

	key := exprKey(assign.Lhs[0])
	if key != "" {
		reserved[key] = true
	}
}

func updateReservedFromDecl(decl *ast.DeclStmt, reserved map[string]bool) {
	gen, ok := decl.Decl.(*ast.GenDecl)
	if !ok || gen.Tok != token.VAR {
		return
	}

	for _, spec := range gen.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		if len(valueSpec.Values) != len(valueSpec.Names) {
			continue
		}
		for i, value := range valueSpec.Values {
			call := extractCall(value, "make")
			if call == nil || !isPreallocMake(call) {
				continue
			}
			name := valueSpec.Names[i]
			if name != nil && name.Name != "" {
				reserved[name.Name] = true
			}
		}
	}
}

func extractCall(expr ast.Expr, name string) *ast.CallExpr {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != name {
		return nil
	}
	return call
}

func isPreallocMake(call *ast.CallExpr) bool {
	argLen := len(call.Args)
	if argLen < 2 {
		return false
	}
	if argLen >= 3 {
		return true
	}
	// len == 2, treat as preallocated when second arg is non-zero literal or non-literal.
	if argLen == 2 {
		if lit, ok := call.Args[1].(*ast.BasicLit); ok && lit.Kind == token.INT && lit.Value == "0" {
			return false
		}
		return true
	}
	return false
}

func cloneReserved(src map[string]bool) map[string]bool {
	if len(src) == 0 {
		return make(map[string]bool, 4)
	}
	out := make(map[string]bool, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func exprKey(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return types.ExprString(e)
	default:
		return types.ExprString(expr)
	}
}

func isSliceRange(info *types.Info, expr ast.Expr) bool {
	if info == nil {
		return false
	}
	t := info.TypeOf(expr)
	if t == nil {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Array, *types.Slice:
		return true
	default:
		return false
	}
}

func isSliceType(info *types.Info, expr ast.Expr) bool {
	if info == nil {
		return false
	}
	t := info.TypeOf(expr)
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Slice)
	return ok
}

func asBlock(stmt ast.Stmt) *ast.BlockStmt {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return s
	case *ast.IfStmt:
		return s.Body
	default:
		return nil
	}
}
