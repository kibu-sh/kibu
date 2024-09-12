package requestrules

import (
	"github.com/kibu-sh/kibu/pkg/wiretap/compare"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/internaltools"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"io"
	"net/http"
)

type MatchFunc func(req *http.Request) (match bool, err error)

var _ spec.MatchRule = MatchRule{}

type MatchRule struct {
	ref        *spec.SnapshotRef
	matchFuncs []MatchFunc
	strategy   func(req *http.Request, matchFuncs []MatchFunc) (match bool, err error)
}

func (r MatchRule) Match(req *http.Request) (match bool, err error) {
	return r.strategy(req, r.matchFuncs)
}

func (r MatchRule) WithHeader(key string, compare compare.Func[string]) MatchRule {
	r.matchFuncs = append(r.matchFuncs, func(req *http.Request) (bool, error) {
		return compare(req.Header.Get(key))
	})
	return r
}

func (r MatchRule) WithQueryParam(key string, compare compare.Func[string]) MatchRule {
	r.matchFuncs = append(r.matchFuncs, func(req *http.Request) (bool, error) {
		return compare(req.URL.Query().Get(key))
	})
	return r
}

func (r MatchRule) WithMethod(compare compare.Func[string]) MatchRule {
	r.matchFuncs = append(r.matchFuncs, func(req *http.Request) (bool, error) {
		return compare(req.Method)
	})
	return r
}

func (r MatchRule) WithPath(compare compare.Func[string]) MatchRule {
	r.matchFuncs = append(r.matchFuncs, func(req *http.Request) (bool, error) {
		return compare(req.URL.Path)
	})
	return r
}

func (r MatchRule) WithHost(compare compare.Func[string]) MatchRule {
	r.matchFuncs = append(r.matchFuncs, func(req *http.Request) (bool, error) {
		return compare(req.Host)
	})
	return r
}

func (r MatchRule) WithBody(contains compare.Func[io.Reader]) MatchRule {
	r.matchFuncs = append(r.matchFuncs, func(req *http.Request) (bool, error) {
		clone, err := internaltools.CloneRequestWithBody(req)
		if err != nil {
			return false, err
		}
		return contains(clone.Body)
	})
	return r
}

func MatchAllStrategy(req *http.Request, matchFuncs []MatchFunc) (match bool, err error) {
	for _, matchFunc := range matchFuncs {
		match, err = matchFunc(req)
		if err != nil || !match {
			return
		}
	}
	return true, nil
}

func NewMatchRule() MatchRule {
	return MatchRule{strategy: MatchAllStrategy}
}

func BasicMatchRule(snapshot *spec.Snapshot) MatchRule {
	return NewMatchRule().
		WithPath(compare.Exactly(snapshot.Request.URL.Path)).
		WithHost(compare.Exactly(snapshot.Request.URL.Host)).
		WithMethod(compare.Exactly(snapshot.Request.Method))
}
