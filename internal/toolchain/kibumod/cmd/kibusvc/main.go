package main

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(kibumod.Analyzer)
}
