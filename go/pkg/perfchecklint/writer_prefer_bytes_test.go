package perfchecklint

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func TestWriterPreferBytesFlagsConversion(t *testing.T) {
	src := `package sample

import (
	"io"
	"strings"
)

func write(w io.Writer, data []byte) {
	io.WriteString(w, string(data))
}

func writeBuilder(data []byte) string {
	var b strings.Builder
	b.WriteString(string(data))
	return b.String()
}
`
	diags := runWriterPreferBytes(src)
	require.Len(t, diags, 2)
	require.True(t, containsRule(diags, "perf_writer_prefer_bytes"))
}

func TestWriterPreferBytesAllowsBytes(t *testing.T) {
	src := `package sample

import (
	"io"
	"strings"
)

func write(w io.Writer, data []byte) {
	w.Write(data)
}

func writeBuilder(data []byte) string {
	var b strings.Builder
	b.Write(data)
	return b.String()
}
`
	diags := runWriterPreferBytes(src)
	require.Empty(t, diags)
}

func runWriterPreferBytes(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "writer.go", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("sample", fset, []*ast.File{file}, info)
	if err != nil {
		panic(err)
	}

	var diags []analysis.Diagnostic
	insp := inspector.New([]*ast.File{file})
	pass := analysis.Pass{
		Analyzer:  writerPreferBytesAnalyzer,
		Fset:      fset,
		Files:     []*ast.File{file},
		Pkg:       pkg,
		TypesInfo: info,
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
		ResultOf: map[*analysis.Analyzer]any{
			inspect.Analyzer: insp,
		},
	}

	if _, err := writerPreferBytesAnalyzer.Run(&pass); err != nil {
		panic(err)
	}

	return diags
}
