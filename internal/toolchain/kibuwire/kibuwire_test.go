package kibuwire

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	results := analysistest.Run(t, testdata,
		Analyzer, "./...")

	providers, ok := results[0].Result.(ProviderList)
	require.True(t, ok)
	require.NotNil(t, providers)
	require.Equal(t, 3, providers.Len())

	grouped := providers.GroupBy(GroupByFQN())
	require.Equal(t, grouped.Len(), 1, "should have 1 group")
	httpHandlers, ok := grouped.Get("github.com/kibu-sh/kibu/pkg/transport/httpx.HandlerFactory")
	require.True(t, ok)
	require.NotNil(t, httpHandlers)
	require.Equal(t, httpHandlers.Len(), 1)

	require.Equal(t, httpHandlers[0].Group.Name, "HandlerFactory")
	require.Equal(t, httpHandlers[0].Group.Import, "github.com/kibu-sh/kibu/pkg/transport/httpx")
}
