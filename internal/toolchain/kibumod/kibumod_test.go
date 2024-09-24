package kibumod

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	tests := []string{"./..."}
	analysistest.Run(t, testdata, Analyzer, tests...)
}
