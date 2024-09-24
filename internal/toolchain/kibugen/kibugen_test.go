package kibugen

import (
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	tests := []string{"./..."}
	err := pipeline.Main(pipeline.Config{
		Patterns:         tests,
		Dir:              testdata,
		Analyzers:        []*analysis.Analyzer{Analyzer},
		FactStore:        pipeline.NoOpFactStore{},
		LoaderConfig:     pipeline.TestingConfig(testdata),
		RunDespiteErrors: true,
	})
	require.NoError(t, err)
}
