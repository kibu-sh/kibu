package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestJSONRequest(t *testing.T) {

	type response struct {
		Message string `json:"message"`
	}
	var err error

	params := JSONRequestParams{
		Method: "POST",
		Client: http.DefaultClient,
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: struct {
			Name string `json:"name"`
		}{
			Name: "client",
		},
	}

	expected := struct {
		Message string `json:"message"`
	}{
		Message: "Hello, client!",
	}

	// create a test server to respond to requests
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return
		}
		var reqBody struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_, err := fmt.Fprintf(w, `{"message": "Hello, %s!"}`, reqBody.Name)
		if err != nil {
			http.Error(w, "no response from server", http.StatusInternalServerError)
			return
		}
	}))
	defer testServer.Close()

	params.Url, err = url.Parse(testServer.URL)
	require.NoError(t, err)

	t.Run("valid request", func(t *testing.T) {

		result, err := JSONRequest[response](context.Background(), params)
		if !reflect.DeepEqual(result.Message, expected.Message) {
			t.Errorf("unexpected result: got %v, want %v", result, expected)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

	})
	t.Run("invalid method on request, use custom status check and error message", func(t *testing.T) {

		params.Method = http.MethodGet
		params.StatusCheck = func(statusCode int, status string) (err error) {
			if statusCode == http.StatusMethodNotAllowed {
				err = errors.New("method is not allowed")
			}
			return
		}
		expectedErrMsg := "method is not allowed"
		_, err := JSONRequest[response](context.Background(), params)
		if err.Error() != expectedErrMsg {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("invalid method on request,when custom status check is nil ,use the default status check and err message", func(t *testing.T) {

		expectedErrMsg := "failed with status:405 Method Not Allowed:code(405): unable to send http request"
		params.Method = http.MethodGet
		params.StatusCheck = nil
		_, err = JSONRequest[response](context.Background(), params)
		if err.Error() != expectedErrMsg {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("when client is nil,should use default client", func(t *testing.T) {

		params.Client = nil
		params.Method = http.MethodPost
		result, err := JSONRequest[response](context.Background(), params)
		if !reflect.DeepEqual(result.Message, expected.Message) {
			t.Errorf("unexpected result: got %v, want %v", result, expected)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("invalid request, wrong header", func(t *testing.T) {

		params.Client = nil
		params.Method = http.MethodPost
		params.Headers["Content-Type"] = []string{"test"}
		expMsg := "failed with status:415 Unsupported Media Type:code(415): unable to send http request"
		_, err := JSONRequest[response](context.Background(), params)
		if err.Error() != expMsg {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("valid request. wrong custom struct", func(t *testing.T) {

		params.Headers["Content-Type"] = []string{"application/json"}
		type res struct {
			Message string `json:"cmsg"`
		}
		result, err := JSONRequest[res](context.Background(), params)
		require.NoError(t, err)
		if result.Message != "" {
			t.Errorf("unexpected error %v", result.Message)
		}

	})
	t.Run("invalid request,wrong body ", func(t *testing.T) {

		params.Client = nil
		params.Method = http.MethodPost

		params.Body = nil
		expMsg := "failed with status:400 Bad Request:code(400): unable to send http request"
		_, err := JSONRequest[response](context.Background(), params)
		if err.Error() != expMsg {
			t.Errorf("unexpected error: %v", err)
		}
	})

}
