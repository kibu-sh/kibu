package archive

import (
	"bufio"
	"bytes"
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/txtar"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
)

var _ spec.SnapshotStore = (*SnapshotArchiveStore)(nil)

const (
	archiveExtension   = ".txtar"
	archiveRequestKey  = "request.bin.http"
	archiveResponseKey = "response.bin.http"
	archiveErrorKey    = "error.bin.txt"
)

type SnapshotArchiveStore struct {
	dir string
}

func NewSnapshotArchiveStore(dir string) SnapshotArchiveStore {
	return SnapshotArchiveStore{dir: dir}
}

func (s SnapshotArchiveStore) Read(ref *spec.SnapshotRef) (snapshot *spec.Snapshot, err error) {
	archiveBytes, err := os.ReadFile(s.snapshotFilename(ref))
	if err != nil {
		return
	}
	return parseRoundTripTxtArchive(archiveBytes, ref.ID)
}

func (s SnapshotArchiveStore) snapshotFilename(ref *spec.SnapshotRef) string {
	return Filename(s.dir, ref)
}

func Filename(dir string, ref *spec.SnapshotRef) string {
	return filepath.Join(dir, ref.ID+archiveExtension)
}

func (s SnapshotArchiveStore) Write(snapshot *spec.Snapshot) (ref *spec.SnapshotRef, err error) {
	ref = snapshot.Ref()
	archive, err := createTxtArchiveFromSnapshot(snapshot)
	if err != nil {
		return
	}
	err = os.WriteFile(s.snapshotFilename(ref), txtar.Format(archive), 0644)
	return
}

func parseRoundTripTxtArchive(archiveBytes []byte, id string) (record *spec.Snapshot, err error) {
	archive, err := safelyParseTxtArchive(archiveBytes)
	if err != nil {
		return
	}

	reqFile, err := getFileFromArchive(archiveRequestKey, archive)
	if err != nil {
		return
	}

	req, err := http.ReadRequest(
		bufio.NewReader(bytes.NewBuffer(reqFile.Data)),
	)
	if err != nil {
		return
	}

	//TODO: make sure this is tested
	//record.Request.URL.Host = req.Host

	resFile, err := getFileFromArchive(archiveResponseKey, archive)
	if err != nil {
		return
	}

	res, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(resFile.Data)), nil)
	if err != nil {
		return
	}

	record, err = spec.NewSnapshot(
		spec.MultiReadRequestCloner{Req: req},
		spec.MultiReadResponseCloner{Res: res},
		0,
	)
	if err != nil {
		return
	}

	// restore original id
	record.ID = strings.TrimSuffix(id, archiveExtension)

	return
}

func safelyParseTxtArchive(archiveBytes []byte) (archive *txtar.Archive, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrapf(err, "panic parsing txt archive: %v", r)
		}
	}()
	archive = txtar.Parse(archiveBytes)
	return
}

func getFileFromArchive(name string, archive *txtar.Archive) (match txtar.File, err error) {
	for _, file := range archive.Files {
		if file.Name == name {
			match = file
			return
		}
	}
	err = errors.Wrapf(spec.ErrFileNotFoundInTxtArchive, "%s", name)
	return
}

func createTxtArchiveFromSnapshot(snapshot *spec.Snapshot) (archive *txtar.Archive, err error) {
	if snapshot.ID == "" {
		err = spec.ErrSnapshotIDRequired
		return
	}

	archive = new(txtar.Archive)

	req, err := spec.DeserializeRequest(snapshot.Request)
	if err != nil {
		return
	}

	// TODO: test this
	// We should never write request headers to disk as they contain sensitive information
	// They're also not necessarily usefully for request matching
	req.Header = copySafeHeaders(snapshot.Request.Header.Clone())

	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return
	}

	res, err := spec.DeserializeResponse(snapshot.Response, nil)
	if err != nil {
		return
	}

	// TODO: test this
	// writing the content-length is unnecessary
	// it also causes issues with templates since the snapshot on disk
	// is modifiable by the user the length can change
	res.ContentLength = -1
	res.Header.Del("Content-Length")
	resDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return
	}

	archive.Files = append(archive.Files, txtar.File{
		Name: archiveRequestKey,
		Data: reqDump,
	})

	archive.Files = append(archive.Files, txtar.File{
		Name: archiveResponseKey,
		Data: resDump,
	})

	return
}

func copySafeHeaders(source http.Header) http.Header {
	clone := make(http.Header)
	for _, key := range safeHeaders {
		for _, val := range source.Values(key) {
			clone.Add(key, val)
		}
	}
	return clone
}

var safeHeaders = []string{
	"Accept",
	"Accept-Encoding",
	"Accept-Language",
	"Cache-Control",
	"Connection",
	// NOT SAFE since responses can be templates which change their original content length
	//"Content-Length",
	"Content-Type",
	"Host",
	"Origin",
	"Referer",
	"User-Agent",
	"X-Test-Header",
	// all grpc headers
	"grpc-accept-encoding",
	"grpc-encoding",
	"grpc-timeout",
	"grpc-trace-bin",
	"grpc-status",
	"grpc-message",
	"grpc-status-details-bin",
	"grpc-previous-rpc-attempts",
	"grpc-retry-pushback-ms",
	"grpc-retry-attempts",
	"grpc-retry-max-attempts",
	"grpc-timeout",
}
