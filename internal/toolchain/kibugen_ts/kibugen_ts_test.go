package kibugen_ts

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analyzerPath := filepath.Join(testdata, "analyzer")
	results := analysistest.Run(t, analyzerPath,
		Analyzer, "./...")

	for _, result := range results {
		if result.Result != nil {
			t.Logf("Generated TypeScript for %s:\n%v", result.Pass.Pkg.Path(), result.Result)
		}
	}
}
