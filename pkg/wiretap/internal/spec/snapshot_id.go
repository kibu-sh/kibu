package spec

import (
	"fmt"
	"github.com/matoous/go-nanoid"
	"github.com/samber/lo"
	"net/url"
	"strings"
	"time"
)

var defaultSnapshotIDFunc = newCanonicalSnapshotIDFunc()

type snapshotIDFunc func(method string, u *url.URL) string

type canonicalSnapshotIDGenerator struct {
	timeGenFunc func() time.Time
	idGenFunc   func() string
}

func newStaticCanonicalIDFunc() snapshotIDFunc {
	return canonicalSnapshotIDGenerator{
		timeGenFunc: func() time.Time {
			return time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC)
		},
		idGenFunc: func() string {
			return "d3f4gx2"
		},
	}.generate
}

func newCanonicalSnapshotIDFunc() snapshotIDFunc {
	return canonicalSnapshotIDGenerator{
		timeGenFunc: time.Now,
		idGenFunc:   nanoidGenFunc(),
	}.generate
}

var nanoAlphabet = "abcdefghijklmnopqrstuvwxyz1234567890"

func nanoidGenFunc() func() string {
	return func() string {
		return gonanoid.MustGenerate(nanoAlphabet, 7)
	}
}

func (c canonicalSnapshotIDGenerator) generate(method string, u *url.URL) string {
	return fmt.Sprintf("%s-%s-%s",
		c.timeGenFunc().Format(time.DateOnly),
		c.idGenFunc(),
		createRequestID(method, u),
	)
}

func createRequestID(method string, u *url.URL) string {
	return strings.Join(lo.Filter([]string{
		strings.ToLower(method),
		strings.ReplaceAll(
			strings.ReplaceAll(u.Host, ".", "_"), ":", "_"),
		strings.ReplaceAll(strings.TrimPrefix(u.Path, "/"), "/", "_"),
	}, filterEmptyStrings), "_")
}

func filterEmptyStrings(item string, index int) bool {
	return item != ""
}
