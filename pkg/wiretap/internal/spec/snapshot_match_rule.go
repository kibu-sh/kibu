package spec

import "net/http"

type MatchRule interface {
	Match(req *http.Request) (match bool, err error)
}
