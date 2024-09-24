package main

import (
	"generatev1/internal/kibusvc"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(kibusvc.Analyzer)
}
