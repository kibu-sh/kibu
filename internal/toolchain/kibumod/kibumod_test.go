package kibumod

import (
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
	result := results[0].Result.(*Package)
	require.NotNil(t, result)
	return
}
