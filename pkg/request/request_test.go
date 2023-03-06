package request

import (
	"context"
	"encoding/json"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestRequest(t *testing.T) {
	ctx := context.Background()
	type message struct {
		Message string `json:"message"`
	}

	type echo struct {
		Data any `json:"data"`
		Code int `json:"code"`
	}

	httpmock.Activate()
	httpmock.RegisterResponder("GET", "https://test.local:8080",
		httpmock.NewStringResponder(200, `{"message": "Hello, client!"}`))

	httpmock.RegisterResponder("POST", "https://test.local:8080", func(request *http.Request) (*http.Response, error) {
		input := new(echo)
		if err := json.NewDecoder(request.Body).Decode(input); err != nil {
			return httpmock.NewJsonResponse(400, message{
				Message: err.Error(),
			})
		}
		return httpmock.NewJsonResponse(input.Code, input.Data)
	})
	defer httpmock.DeactivateAndReset()

	client, parseErr := ParseURL("https://test.local:8080")
	require.NoError(t, parseErr)

	t.Run("should build a client with default options", func(t *testing.T) {
		require.NotNil(t, client.c)
		require.NotNil(t, client.baseURL)
		require.Equal(t, client.body, http.NoBody)
		require.NotNil(t, client.statusCheckFunc)
		require.NotNil(t, client.defaultHeader)
		require.NoError(t, client.statusCheckFunc("OK", 200))
		require.NoError(t, client.statusCheckFunc("OK", 201))
		require.ErrorIs(t, client.statusCheckFunc("OK", 400), ErrStatusCheckFailed)
	})

	t.Run("should properly execute basic json request", func(t *testing.T) {
		msg := new(message)
		err := client.DoAsJSON(ctx, msg)
		require.NoError(t, err)
		require.Equal(t, "Hello, client!", msg.Message)
	})

	t.Run("should send and receive request as JSON", func(t *testing.T) {
		c := client.
			WithPost().
			WithStatusCheckFunc(
				NewExactStatusCheckFunc(201),
			).
			WithJSONBody(echo{
				Code: 201,
				Data: message{
					Message: "Hello, server!",
				},
			})

		require.NotNil(t, c.deferredBody)
		require.Equal(
			t,
			c.defaultHeader.Get("Content-Type"),
			"application/json",
			"content type should be automatically set to application/json",
		)

		msg := new(message)
		err := c.DoAsJSON(ctx, msg)
		require.NoError(t, err)
		require.Equal(t, "Hello, server!", msg.Message)
	})
}
