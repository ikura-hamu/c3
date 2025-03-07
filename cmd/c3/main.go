package main

import (
	"github.com/ikura-hamu/c3"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(c3.Analyzer) }
