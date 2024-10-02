package dynamic

import (
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"net/http"
)

var _ spec.SnapshotRouter = &SnapshotRouter{}

type SnapshotRouter struct {
	RuleGroups    []spec.RuleGroup
	MatchStrategy spec.SnapshotMatchFunc
}

func (s *SnapshotRouter) Register(ref *spec.SnapshotRef, rules ...spec.MatchRule) spec.SnapshotRouter {
	s.RuleGroups = append(s.RuleGroups, spec.RuleGroup{
		Ref:        ref,
		MatchRules: rules,
	})
	return s
}

func (s *SnapshotRouter) Match(req *http.Request) (ref *spec.SnapshotRef, err error) {
	return s.MatchStrategy(req, s.RuleGroups)
}

func NewSnapshotRouter() *SnapshotRouter {
	return &SnapshotRouter{
		MatchStrategy: RequireAllRulesMatchStrategy,
	}
}

func RequireAllRulesMatchStrategy(req *http.Request, groups []spec.RuleGroup) (ref *spec.SnapshotRef, err error) {
	for _, group := range groups {
		match := true
		for _, rule := range group.MatchRules {
			match, err = rule.Match(req)
			if err != nil {
				return nil, err
			}
			if !match {
				break
			}
		}
		if match {
			return group.Ref, nil
		}
	}
	return nil, nil
}
