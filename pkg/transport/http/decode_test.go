package http

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestDecodeForm(t *testing.T) {
	type payload struct {
		Foo    string
		Json   string `json:"foo"`
		Query  string `query:"foo"`
		Header string `header:"Host"`
		Cookie string `cookie:"foo"`
		Nested struct {
			Name   string `query:"foo"`
			Number int    `query:"bar"`
		}
	}

	example := payload{}

	u, err := url.Parse("https://example.com?foo=bar&bar=1")
	require.NoError(t, err)

	req := &http.Request{
		URL:    u,
		Method: http.MethodPost,
		Header: http.Header{
			"Host": []string{"example.com"},
			"Content-Type": []string{
				"application/json; charset=utf-8",
			},
		},
		Body: io.NopCloser(bytes.NewBuffer([]byte(`{
			"foo": "bar",
			"Foo": "bar"
		}`))),
		// TODO: clean up tests and check for form cases
		//Form: url.Values{
		//	"foo":    []string{"bar"},
		//	"number": []string{"1"},
		//},
		//PostForm: url.Values{
		//	"foo": []string{"baz"},
		//},
		//MultipartForm: &multipart.Form{
		//	Value: url.Values{
		//		"foo": []string{"qux"},
		//	},
		//	File: nil,
		//},
	}

	req.AddCookie(&http.Cookie{
		Name:  "foo",
		Value: "bar",
	})

	err = Decode[payload](req)(context.Background(), &example)
	require.NoError(t, err)

	require.Equal(t, "bar", example.Foo)
	require.Equal(t, "bar", example.Json)
	require.Equal(t, req.Header.Get("host"), example.Header)
	require.Equal(t, req.URL.Query().Get("foo"), example.Query)
	require.Equal(t, "bar", example.Cookie)
	require.Equal(t, "bar", example.Nested.Name)
	require.Equal(t, 1, example.Nested.Number)
}
