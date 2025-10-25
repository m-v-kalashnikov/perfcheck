package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

var bufferedIOAnalyzer = &analysis.Analyzer{
	Name:     "perf_use_buffered_io",
	Doc:      "reports repeated small I/O without buffering",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		rule, ok := ruleset.MustDefault().RuleByID("perf_use_buffered_io")
		if !ok {
			return nil, fmt.Errorf("rule perf_use_buffered_io not found")
		}

		ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if ins == nil {
			return nil, fmt.Errorf("missing inspector dependency")
		}

		nodeFilter := []ast.Node{(*ast.ForStmt)(nil), (*ast.RangeStmt)(nil)}
		ins.Preorder(nodeFilter, func(node ast.Node) {
			var body *ast.BlockStmt
			switch loop := node.(type) {
			case *ast.ForStmt:
				body = loop.Body
			case *ast.RangeStmt:
				body = loop.Body
			}
			inspectLoopIO(pass, body, rule)
		})

		return nil, nil
	},
}

func inspectLoopIO(pass *analysis.Pass, body *ast.BlockStmt, rule ruleset.Rule) {
	if body == nil {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch {
		case isFmtPrintCall(pass, call):
			report(pass, call.Lparen, rule, "fmt.Fprint inside loop performs unbuffered writes")
		case isIOWriteStringCall(pass, call):
			report(
				pass,
				call.Lparen,
				rule,
				"io.WriteString inside loop is unbuffered; wrap the writer with bufio",
			)
		case isSmallUnbufferedWrite(pass, call):
			report(pass, call.Lparen, rule, "loop writes tiny byte slices without buffering")
		}
		return true
	})
}

func isFmtPrintCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || pass.TypesInfo == nil {
		return false
	}
	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != "fmt" {
		return false
	}
	name := obj.Name()
	if name != "Fprint" && name != "Fprintf" && name != "Fprintln" {
		return false
	}
	if len(call.Args) == 0 {
		return false
	}
	if isBufferedWriterType(pass.TypesInfo.TypeOf(call.Args[0])) {
		return false
	}
	return true
}

func isIOWriteStringCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || pass.TypesInfo == nil {
		return false
	}
	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != "io" {
		return false
	}
	if obj.Name() != "WriteString" {
		return false
	}
	if len(call.Args) < 1 {
		return false
	}
	if isBufferedWriterType(pass.TypesInfo.TypeOf(call.Args[0])) {
		return false
	}
	return true
}

func isSmallUnbufferedWrite(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || pass.TypesInfo == nil {
		return false
	}
	selInfo := pass.TypesInfo.Selections[sel]
	method := sel.Sel.Name
	if method != "Write" && method != "WriteByte" && method != "WriteRune" {
		return false
	}
	var recv types.Type
	if selInfo != nil {
		recv = selInfo.Recv()
	} else {
		recv = pass.TypesInfo.TypeOf(sel.X)
	}
	if isBufferedWriterType(recv) {
		return false
	}
	if method == "WriteByte" || method == "WriteRune" {
		return true
	}
	if len(call.Args) == 0 {
		return false
	}
	return isSmallBytePayload(pass, call.Args[0])
}

func isBufferedWriterType(typ types.Type) bool {
	if typ == nil {
		return false
	}
	switch t := typ.(type) {
	case *types.Pointer:
		return isBufferedWriterType(t.Elem())
	case *types.Named:
		obj := t.Obj()
		if obj == nil || obj.Pkg() == nil {
			return false
		}
		pkg := obj.Pkg().Path()
		if pkg == "bufio" {
			return obj.Name() == "Writer" || obj.Name() == "Reader"
		}
		if pkg == "bytes" && obj.Name() == "Buffer" {
			return true
		}
		return false
	default:
		return isBufferedWriterType(t.Underlying())
	}
}

func isSmallBytePayload(pass *analysis.Pass, expr ast.Expr) bool {
	typ := pass.TypesInfo.TypeOf(expr)
	if typ == nil {
		return false
	}
	if slice, ok := typ.Underlying().(*types.Slice); ok {
		if basic, ok := slice.Elem().(*types.Basic); !ok || basic.Kind() != types.Byte {
			return false
		}
	} else {
		return false
	}
	switch v := expr.(type) {
	case *ast.CompositeLit:
		return len(v.Elts) > 0 && len(v.Elts) <= 4
	case *ast.CallExpr:
		return isSmallMakeSlice(v)
	case *ast.BasicLit:
		if v.Kind == token.STRING {
			if unquoted, err := strconv.Unquote(v.Value); err == nil {
				return len([]rune(unquoted)) > 0 && len([]rune(unquoted)) <= 4
			}
		}
	}
	return false
}

func isSmallMakeSlice(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "make" {
		return false
	}
	if len(call.Args) < 2 {
		return false
	}
	typ, ok := call.Args[0].(*ast.ArrayType)
	if !ok {
		return false
	}
	elt, ok := typ.Elt.(*ast.Ident)
	if !ok || elt.Name != "byte" {
		return false
	}
	sizeLit, ok := call.Args[1].(*ast.BasicLit)
	if !ok || sizeLit.Kind != token.INT {
		return false
	}
	return parseSmallInt(sizeLit.Value)
}

func parseSmallInt(text string) bool {
	var value int
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return false
		}
		value = value*10 + int(ch-'0')
		if value > 4 {
			return false
		}
	}
	return value > 0 && value <= 4
}
