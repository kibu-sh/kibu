package kibudecor

import (
	"encoding/gob"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
	"os"
	"path/filepath"
	"testing"
)

func gobFile(dir string) string {
	return filepath.Join(dir, "expected.gob")
}

func saveGobData[T any](dir string, result T) (err error) {
	file, err := os.OpenFile(gobFile(dir),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return
	}

	err = gob.NewEncoder(file).Encode(result)
	return
}

func loadGobData[T any](dir string) (result T, err error) {
	file, err := os.Open(gobFile(dir))
	if err != nil {
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	err = gob.NewDecoder(file).Decode(&result)
	return
}

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	results := analysistest.Run(t, testdata,
		Analyzer, "./...")

	values := results[0].Result.(*Map)
	var directives decorators.List
	for dir := range values.ValuesFromOldest() {
		directives = append(directives, dir...)
	}

	err := saveGobData(testdata, directives)
	require.NoError(t, err)

	loaded, err := loadGobData[decorators.List](testdata)
	require.NoError(t, err)
	require.NotNil(t, directives)
	require.Equal(t, directives, loaded)

	hasKibuService := directives.Some(decorators.HasKey("kibu", "service"))
	require.True(t, hasKibuService, "should have kibu:service")
}
