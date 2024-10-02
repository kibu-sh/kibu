package spec

import (
	"net/http"
)

type SnapshotRouter interface {
	Match(req *http.Request) (ref *SnapshotRef, er error)
	Register(ref *SnapshotRef, rules ...MatchRule) SnapshotRouter
}

type RuleGroup struct {
	Ref        *SnapshotRef
	MatchRules []MatchRule
}

type SnapshotMatchFunc func(req *http.Request, groups []RuleGroup) (ref *SnapshotRef, err error)
