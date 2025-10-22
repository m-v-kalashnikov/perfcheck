package main

import (
	"github.com/m-v-kalashnikov/perfcheck/go/internal/analyzer"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(analyzer.All()...)
}
