package main

import (
	"github.com/gostaticanalysis/typednil"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(typednil.Analyzer) }
