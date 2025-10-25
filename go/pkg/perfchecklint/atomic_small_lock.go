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

var atomicSmallLockAnalyzer = &analysis.Analyzer{
	Name:     "perf_atomic_for_small_lock",
	Doc:      "reports mutexes guarding single primitive values",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_atomic_for_small_lock")
		if !ok {
			return nil, fmt.Errorf("rule perf_atomic_for_small_lock not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		nodeFilter := []ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}
		ins.Preorder(nodeFilter, func(node ast.Node) {
			var body *ast.BlockStmt
			switch fn := node.(type) {
			case *ast.FuncDecl:
				body = fn.Body
			case *ast.FuncLit:
				body = fn.Body
			}
			analyzeMutexSections(pass, body, rule)
		})

		return nil, nil
	},
}

type mutexTarget struct {
	expr string
}

func analyzeMutexSections(pass *analysis.Pass, body *ast.BlockStmt, rule ruleset.Rule) {
	if body == nil {
		return
	}
	stmts := body.List
	for i := 0; i < len(stmts); i++ {
		receiver, ok := lockCallReceiver(pass, stmts[i])
		if !ok {
			continue
		}

		if i+2 >= len(stmts) {
			continue
		}
		subject, ok := primitiveMutation(pass, stmts[i+1])
		if !ok {
			continue
		}
		if !isUnlockCall(pass, stmts[i+2], receiver) {
			continue
		}
		msg := fmt.Sprintf("mutex %s guards primitive %s; use sync/atomic", receiver.expr, subject)
		report(pass, stmts[i+1].Pos(), rule, msg)
		i += 2
	}
}

func lockCallReceiver(pass *analysis.Pass, stmt ast.Stmt) (mutexTarget, bool) {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return mutexTarget{}, false
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return mutexTarget{}, false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return mutexTarget{}, false
	}
	selInfo := pass.TypesInfo.Selections[sel]
	if selInfo == nil {
		return mutexTarget{}, false
	}
	if sel.Sel == nil || sel.Sel.Name != "Lock" {
		return mutexTarget{}, false
	}
	if !isSyncMutex(selInfo.Recv()) {
		return mutexTarget{}, false
	}
	return makeMutexTarget(pass, sel.X), true
}

func isUnlockCall(pass *analysis.Pass, stmt ast.Stmt, receiver mutexTarget) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil {
		return false
	}
	name := sel.Sel.Name
	if name != "Unlock" && name != "RUnlock" {
		return false
	}
	other := makeMutexTarget(pass, sel.X)
	if !sameReceiver(receiver, other) {
		return false
	}
	selInfo := pass.TypesInfo.Selections[sel]
	return selInfo == nil || isSyncMutex(selInfo.Recv())
}

func makeMutexTarget(pass *analysis.Pass, expr ast.Expr) mutexTarget {
	return mutexTarget{expr: types.ExprString(expr)}
}

func sameReceiver(a, b mutexTarget) bool {
	return a.expr == b.expr
}

func primitiveMutation(pass *analysis.Pass, stmt ast.Stmt) (string, bool) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if len(s.Lhs) != 1 || len(s.Rhs) != 1 {
			return "", false
		}
		lhs := s.Lhs[0]
		if !isSimpleExpr(lhs) {
			return "", false
		}
		if !isPrimitiveType(pass.TypesInfo.TypeOf(lhs)) {
			return "", false
		}
		if s.Tok != token.ASSIGN && s.Tok != token.ADD_ASSIGN && s.Tok != token.SUB_ASSIGN && s.Tok != token.DEFINE {
			return "", false
		}
		return types.ExprString(lhs), true
	case *ast.IncDecStmt:
		if !isSimpleExpr(s.X) {
			return "", false
		}
		if !isPrimitiveType(pass.TypesInfo.TypeOf(s.X)) {
			return "", false
		}
		return types.ExprString(s.X), true
	default:
		return "", false
	}
}

func isSimpleExpr(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

func isSyncMutex(typ types.Type) bool {
	if typ == nil {
		return false
	}
	switch t := typ.(type) {
	case *types.Pointer:
		return isSyncMutex(t.Elem())
	case *types.Named:
		obj := t.Obj()
		if obj == nil || obj.Pkg() == nil {
			return false
		}
		if obj.Pkg().Path() != "sync" {
			return false
		}
		return obj.Name() == "Mutex" || obj.Name() == "RWMutex"
	default:
		return isSyncMutex(t.Underlying())
	}
}

func isPrimitiveType(t types.Type) bool {
	if t == nil {
		return false
	}
	under := t.Underlying()
	switch typ := under.(type) {
	case *types.Basic:
		return isPrimitiveKind(typ.Kind())
	case *types.Pointer:
		if basic, ok := typ.Elem().Underlying().(*types.Basic); ok {
			return isPrimitiveKind(basic.Kind())
		}
	}
	return false
}

func isPrimitiveKind(kind types.BasicKind) bool {
	switch kind {
	case types.Bool,
		types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
		types.Float32, types.Float64,
		types.Complex64, types.Complex128,
		types.UnsafePointer:
		return true
	default:
		return false
	}
}
