package dynamic

import (
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

type mockRule struct {
	mock.Mock
}

func (m *mockRule) Match(req *http.Request) (match bool, err error) {
	args := m.Called(req)
	return args.Bool(0), args.Error(1)
}

func TestSnapshotRouter(t *testing.T) {
	rule := &mockRule{}
	goodRequest := &http.Request{
		Method: http.MethodGet,
		URL:    lo.Must(url.Parse("https://localhost:8080/echo")),
	}

	badRequest := &http.Request{
		Method: http.MethodGet,
		URL:    lo.Must(url.Parse("https://localhost/bad")),
	}

	rule.On("Match",
		mock.MatchedBy(compareRequestBasic(goodRequest)),
	).Return(true, nil)

	rule.On("Match",
		mock.MatchedBy(compareRequestBasic(badRequest)),
	).Return(false, nil)

	router := NewSnapshotRouter().
		Register(spec.NewSnapshotRef("123"), rule)

	ref, err := router.Match(goodRequest)
	require.NoError(t, err)
	require.Equal(t, "123", ref.ID)

	ref, err = router.Match(badRequest)
	require.NoError(t, err)
	require.Nil(t, ref)

	rule.AssertExpectations(t)
}

func compareRequestBasic(expectedRequest *http.Request) func(req *http.Request) bool {
	return func(req *http.Request) bool {
		return req.Method == expectedRequest.Method && req.URL.String() == expectedRequest.URL.String()
	}
}
