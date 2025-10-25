package perfchecklint

import "testing"

func TestBufferedIOAnalyzerFlagsFmtWrites(t *testing.T) {
	src := `package sample

import (
	"fmt"
	"os"
)

func writeAll(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return f.Close()
}
`

	diags := runAnalyzerOnSource(t, bufferedIOAnalyzer, "buffered_io.go", src)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !containsRule(diags, "perf_use_buffered_io") {
		t.Fatalf("missing perf_use_buffered_io diagnostic")
	}
}

func TestBufferedIOAnalyzerIgnoresBufioWrites(t *testing.T) {
	src := `package sample

import (
	"bufio"
	"os"
)

func writeAll(path string, payload []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	buf := bufio.NewWriter(f)
	for _, b := range payload {
		if _, err := buf.Write([]byte{b}); err != nil {
			return err
		}
	}
	return buf.Flush()
}
`

	diags := runAnalyzerOnSource(t, bufferedIOAnalyzer, "buffered_io_ok.go", src)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}
}
