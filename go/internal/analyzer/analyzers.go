// Package analyzer hosts perfcheck-specific static analysis detectors.
package analyzer

import "golang.org/x/tools/go/analysis"

// All returns the analyzers implemented by perfcheck.
func All() []*analysis.Analyzer {
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
