package kibumod

import (
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/stretchr/testify/require"
	"log"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	if err != nil {
		log.Fatal(err)
	}
	tests := []string{"./..."}
	results := analysistest.Run(t, testdata, Analyzer, tests...)
	result := results[0].Result.(*modspecv2.Package)
	require.NotNil(t, result)
	return
}
