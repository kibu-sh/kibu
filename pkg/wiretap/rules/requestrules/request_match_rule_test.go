package requestrules

import (
	"github.com/kibu-sh/kibu/pkg/wiretap/compare"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func TestMachRules(t *testing.T) {
	rule := NewMatchRule().
		WithQueryParam("hello", compare.Contains("echo")).
		WithMethod(compare.Exactly(http.MethodPost)).
		WithPath(compare.HasPrefix("/echo")).
		WithHost(compare.Exactly("localhost:8088")).
		WithPath(compare.Glob("/echo/sub/*")).
		WithBody(compare.JSON("hello", compare.Contains("echo")))

	body := `{"hello": "echo"}`
	bodyBuffer := strings.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, "https://localhost:8088/echo/sub/path?hello=echo", bodyBuffer)
	require.NoError(t, err)

	match, err := rule.Match(req)
	require.NoError(t, err)
	require.Truef(t, match, "expected request to match rule")
}
