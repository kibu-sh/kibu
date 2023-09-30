package spec

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func Test__createRequestID(t *testing.T) {
	tests := []struct {
		method   string
		url      string
		expected string
	}{
		{http.MethodGet, "https://api.example.com", "get_api_example_com"},
		{http.MethodPost, "https://example.com", "post_example_com"},
		{http.MethodPut, "https://example.com", "put_example_com"},
		{http.MethodDelete, "https://example.com", "delete_example_com"},
		{http.MethodPatch, "https://example.com", "patch_example_com"},
		{http.MethodOptions, "https://example.com", "options_example_com"},
		{http.MethodHead, "https://example.com", "head_example_com"},
		{http.MethodTrace, "https://example.com", "trace_example_com"},
		{http.MethodConnect, "https://example.com", "connect_example_com"},
		{http.MethodGet, "https://example.com/path", "get_example_com_path"},
		{http.MethodGet, "https://example.com/path?test=true&search=query", "get_example_com_path"},
		{http.MethodGet, "https://example.com/path/With/Capitalization", "get_example_com_path_With_Capitalization"},
		{http.MethodGet, "https://example.com:9090/path", "get_example_com_9090_path"},
	}

	t.Run("should return the expected archive file name", func(t *testing.T) {
		for _, test := range tests {
			actual := createRequestID(test.method, lo.Must(url.Parse(test.url)))
			require.Equal(t, test.expected, actual)
		}
	})
}

func Test__canonicalSnapshotIDGenerator(t *testing.T) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    lo.Must(url.Parse("https://example.com/path")),
	}
	genID := newStaticCanonicalIDFunc()
	expected := "2021-01-03-d3f4gx2-get_example_com_path"
	require.Equal(t, expected, genID(req.Method, req.URL))

	genID = newCanonicalSnapshotIDFunc()
	actual := genID(req.Method, req.URL)
	require.NotEmpty(t, actual)
	require.Contains(t, actual, "get_example_com_path")
}
