package archive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"github.com/rogpeppe/go-internal/txtar"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func Test__archive__functions(t *testing.T) {
	reqBody := []byte(`{"test": "request"}`)
	resBody := []byte(`{"test": "response"}`)
	cwd, _ := os.Getwd()
	testdata := filepath.Join(cwd, "testdata")
	tmpDir, _ := os.MkdirTemp("", "archive")
	expectedArchiveFile := filepath.Join(testdata, "golden.json.txtar")
	actualArchiveFile := filepath.Join(tmpDir, "golden.json.txtar")
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	reqURL := &url.URL{
		Scheme:   "https",
		Host:     "example.com",
		Path:     "/",
		RawQuery: "test=true",
	}

	request := &http.Request{
		Method:     http.MethodPost,
		Host:       "example.com",
		URL:        reqURL,
		RequestURI: reqURL.String(),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"X-Test-Header":  []string{"request"},
			"Content-Length": []string{fmt.Sprintf("%d", len(reqBody))},
		},
		Body:          io.NopCloser(bytes.NewReader(reqBody)),
		ContentLength: int64(len(reqBody)),
		RemoteAddr:    "0.0.0.0",
	}

	response := &http.Response{
		Status:     "400 bad request",
		StatusCode: http.StatusBadRequest,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"X-Test-Header":  []string{"response"},
			"Content-Length": []string{fmt.Sprintf("%d", len(resBody))},
		},
		Body:          io.NopCloser(bytes.NewReader(resBody)),
		ContentLength: int64(len(resBody)),
		Request:       request,
	}

	expectedSnapshot, shErr := spec.NewSnapshot(
		spec.MultiReadRequestCloner{Req: request},
		spec.MultiReadResponseCloner{Res: response},
		0,
	)
	require.NoError(t, shErr)

	t.Run("should reject snapshot with missing id", func(t *testing.T) {
		_, err := createTxtArchiveFromSnapshot(&spec.Snapshot{})
		require.Error(t, err)
		require.ErrorIs(t, err, spec.ErrSnapshotIDRequired)
	})

	t.Run("should successfully construct a txt archive from a round trip snapshot", func(t *testing.T) {
		archive, err := createTxtArchiveFromSnapshot(expectedSnapshot)
		require.NoError(t, err)

		reqFile, err := getFileFromArchive(archiveRequestKey, archive)
		require.NoError(t, err)
		require.NotNil(t, reqFile)

		resFile, err := getFileFromArchive(archiveResponseKey, archive)
		require.NoError(t, err)
		require.NotNil(t, resFile)

		err = os.WriteFile(actualArchiveFile, txtar.Format(archive), 0644)
		require.NoError(t, err)
	})

	t.Run("should successfully construct a snapshot from a txt archive", func(t *testing.T) {
		archiveBytes, err := os.ReadFile(actualArchiveFile)
		require.NoError(t, err)

		// replace /r/n with /n
		// because my editor keeps changing the line endings
		archiveBytes = bytes.ReplaceAll(archiveBytes, []byte("\r\n"), []byte("\n"))

		expectedArchiveBytes, err := os.ReadFile(expectedArchiveFile)
		require.NoError(t, err)

		require.Equal(t, string(expectedArchiveBytes), string(archiveBytes))

		actualSnapshot, err := parseRoundTripTxtArchive(archiveBytes, expectedSnapshot.ID)
		require.NoError(t, err)

		actualSnapshot.Response.ContentLength = 0
		require.EqualValues(t, expectedSnapshot, actualSnapshot)

		var actualRequestBody map[string]any
		err = json.Unmarshal([]byte(actualSnapshot.Request.Body), &actualRequestBody)
		require.NoError(t, err, "should be able to deserialize expected request body")
		require.Equal(t, "request", actualRequestBody["test"])

		require.EqualValues(t, expectedSnapshot.Response.Header, actualSnapshot.Response.Header)
		require.EqualValues(t, expectedSnapshot.Response.StatusCode, actualSnapshot.Response.StatusCode)
		require.Equal(t, expectedSnapshot.Response.Body, actualSnapshot.Response.Body)

		var actualResponseBody map[string]any
		err = json.Unmarshal([]byte(actualSnapshot.Response.Body), &actualResponseBody)
		require.NoError(t, err, "should be able to deserialize expected response body")
		require.Equal(t, "response", actualResponseBody["test"])

		require.NoError(t, err)
	})
}

func TestSnapshotArchiveStore(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "archive")
	store := NewSnapshotArchiveStore(tmpDir)
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	var ref *spec.SnapshotRef
	t.Run("should successfully write a snapshot to the archive", func(t *testing.T) {
		var err error
		u := lo.Must(url.Parse("https://example.com"))
		r := spec.MultiReadRequestCloner{
			Req: &http.Request{
				Method:     http.MethodGet,
				URL:        u,
				RequestURI: u.String(),
			},
		}
		w := spec.MultiReadResponseCloner{
			Res: &http.Response{
				Status:     "OK",
				StatusCode: http.StatusOK,
			},
		}
		sh, _ := spec.NewSnapshot(r, w, 0)
		ref, err = store.Write(sh)
		require.NoError(t, err)
		require.NotNil(t, ref)
		require.FileExistsf(t, filepath.Join(tmpDir, ref.ID+archiveExtension), "expected snapshot to be written to the archive")
	})

	t.Run("should successfully read a snapshot from the archive", func(t *testing.T) {
		snapshot, err := store.Read(ref)
		require.NoError(t, err)
		require.NotNil(t, snapshot)
		require.Equal(t, http.MethodGet, snapshot.Request.Method)
		require.Equal(t, "https://example.com", snapshot.Request.URL.String())
		require.Equal(t, http.StatusOK, snapshot.Response.StatusCode)
	})
}
