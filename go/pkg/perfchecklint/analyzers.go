// Package perfchecklint hosts perfcheck-specific static analysis detectors.
package perfchecklint

import "golang.org/x/tools/go/analysis"

// Analyzers returns the analyzers implemented by perfcheck.
func Analyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		stringConcatLoopAnalyzer,
		regexCompileLoopAnalyzer,
		preallocateCollectionsAnalyzer,
		reflectionLoopAnalyzer,
		boundConcurrencyAnalyzer,
		equalFoldAnalyzer,
		syncPoolPointerAnalyzer,
		writerPreferBytesAnalyzer,
		linkedListAnalyzer,
		atomicSmallLockAnalyzer,
		deferInLoopAnalyzer,
		runeConversionAnalyzer,
		bufferedIOAnalyzer,
		stackAllocAnalyzer,
	}
}

// All is a deprecated alias for Analyzers kept for transitional callers.
func All() []*analysis.Analyzer {
	return Analyzers()
}
