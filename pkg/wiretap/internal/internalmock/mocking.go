package internalmock

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/messaging"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/url"
)

var _ spec.SnapshotMessageTopic = (*MockSnapshotMessageTopic)(nil)

type MockSnapshotMessageTopic struct {
	mock.Mock
}

func (m *MockSnapshotMessageTopic) Publish(ctx context.Context, message spec.Snapshot) (err error) {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockSnapshotMessageTopic) Subscribe(ctx context.Context) (stream messaging.Stream[spec.Snapshot], err error) {
	args := m.Called(ctx)
	return args.Get(0).(messaging.Stream[spec.Snapshot]), args.Error(1)
}

var _ http.RoundTripper = (*MockRoundTripper)(nil)

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	args := m.Called(request)
	return args.Get(0).(*http.Response), args.Error(1)
}

var _ spec.SnapshotStore = &MockSnapshotStore{}

type MockSnapshotStore struct {
	mock.Mock
}

func (m *MockSnapshotStore) Read(ref *spec.SnapshotRef) (snapshot *spec.Snapshot, err error) {
	args := m.Called(ref)
	return args.Get(0).(*spec.Snapshot), args.Error(1)
}

func (m *MockSnapshotStore) Write(snapshot *spec.Snapshot) (ref *spec.SnapshotRef, err error) {
	args := m.Called(snapshot)
	return args.Get(0).(*spec.SnapshotRef), args.Error(1)
}

var _ spec.SnapshotRouter = &MockSnapshotRouter{}

type MockSnapshotRouter struct {
	mock.Mock
}

func (m *MockSnapshotRouter) Match(req *http.Request) (ref *spec.SnapshotRef, er error) {
	args := m.Called(req)
	return args.Get(0).(*spec.SnapshotRef), args.Error(1)
}

func (m *MockSnapshotRouter) Register(ref *spec.SnapshotRef, rules ...spec.MatchRule) spec.SnapshotRouter {
	args := m.Called(ref, rules)
	return args.Get(0).(spec.SnapshotRouter)
}

func NewTestSnapshotRef() *spec.SnapshotRef {
	return new(spec.SnapshotRef)
}

func NewTestResponse() *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
	}
}

func NewTestRequest() *http.Request {
	return &http.Request{
		Method: http.MethodGet,
		URL:    lo.Must(url.Parse("https://localhost:8080/echo")),
	}
}

func NewTestSerializableRequest() *spec.SerializableRequest {
	r, _ := spec.SerializeRequest(NewTestRequestCloner(NewTestRequest()))
	return r
}

func NewTestSerializableResponse() *spec.SerializableResponse {
	r, _ := spec.SerializeResponse(NewTestResponseCloner(NewTestResponse()))
	return r
}

func NewTestSnapshot() *spec.Snapshot {
	s, _ := spec.NewSnapshot(
		NewTestRequestCloner(NewTestRequest()),
		NewTestResponseCloner(NewTestResponse()), 0)
	return s
}

func NewTestRequestCloner(req *http.Request) spec.RequestCloner {
	return spec.MultiReadRequestCloner{Req: req}
}

func NewTestResponseCloner(res *http.Response) spec.ResponseCloner {
	return spec.MultiReadResponseCloner{
		Res: res,
	}
}
