package spec

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func TestRequestTemplateVars(t *testing.T) {
	var tests = map[string]struct {
		body     string
		template string
		expected string
		header   http.Header
		url      string
	}{
		"should render a template with JSON": {
			url:      "https://example.com",
			body:     `{"hello": "json"}`,
			template: `${{ .Request.JSON "hello" }}`,
			expected: "json",
			header:   http.Header{"Content-Type": []string{"application/json"}},
		},
		"should render a template with json as its contents": {
			url:      "https://example.com",
			body:     `{"hello": "json"}`,
			template: `{"hello":"${{ .Request.JSON "hello" }}"}`,
			expected: `{"hello":"json"}`,
		},
		"should render a template with a header": {
			url:      "https://example.com",
			body:     `{"hello": "json"}`,
			template: `${{ .Request.Header "Content-Type" }}`,
			expected: "application/json",
			header:   http.Header{"Content-Type": []string{"application/json"}},
		},
		"should render a template with a form": {
			url:      "https://example.com",
			body:     url.Values{"hello": []string{"form"}}.Encode(),
			template: `${{ .Request.Form "hello" }}`,
			expected: "form",
			header:   http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		},
		"should render a template with query params": {
			url:      "https://example.com?hello=query",
			template: `${{ .Request.Query "hello" }}`,
			expected: "query",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("POST", test.url,
				bytes.NewBufferString(test.body))
			require.NoError(t, err)
			req.Header = test.header

			templateVars, err := RequestToTemplate(req)
			tmp, err := BodyTemplate().Parse(test.template)
			require.NoError(t, err)

			buf := bytes.NewBuffer(nil)
			err = tmp.Execute(buf, map[string]any{
				"Request": templateVars,
			})
			require.NoError(t, err)
			require.Equal(t, test.expected, buf.String())
		})
	}
}
