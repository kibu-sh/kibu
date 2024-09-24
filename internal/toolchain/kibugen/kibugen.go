package kibugen

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibumod"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name:             "kibugen",
	Doc:              "Analyzes go source code and generates system plumbing code for kibu applications",
	Run:              run,
	RunDespiteErrors: true,
	Requires: []*analysis.Analyzer{
		kibumod.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	return nil, nil
}
