package proxy

import (
	"errors"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/internalmock"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func newMockReplayTransport() (*internalmock.MockSnapshotRouter, *internalmock.MockSnapshotStore, *ReplayTransport) {
	mockStore := new(internalmock.MockSnapshotStore)
	mockRouter := new(internalmock.MockSnapshotRouter)
	player := NewReplayTransport(mockRouter, mockStore)
	return mockRouter, mockStore, player
}

func TestTransportPlayer_RoundTrip(t *testing.T) {
	goodRequest := &http.Request{
		Method: http.MethodGet,
		URL:    lo.Must(url.Parse("https://example.com")),
	}

	t.Run("should return an error if the router returns an error", func(t *testing.T) {
		mockRouter, _, player := newMockReplayTransport()
		mockRouter.On("Match", mock.Anything).Return((*spec.SnapshotRef)(nil), errors.New("failure"))
		_, err := player.RoundTrip(goodRequest)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRouteMatch)
		mockRouter.AssertExpectations(t)
	})

	t.Run("should return an error if the router returns no match", func(t *testing.T) {
		mockRouter, _, player := newMockReplayTransport()
		mockRouter.On("Match", mock.Anything).Return((*spec.SnapshotRef)(nil), nil)
		_, err := player.RoundTrip(goodRequest)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNoRouteMatch)
		mockRouter.AssertExpectations(t)
	})

	t.Run("should return an error if the store returns an error", func(t *testing.T) {
		mockRouter, mockStore, player := newMockReplayTransport()
		mockRouter.On("Match", mock.Anything).Return(&spec.SnapshotRef{}, nil)
		mockStore.On("Read", mock.Anything).Return((*spec.Snapshot)(nil), errors.New("failure"))
		_, err := player.RoundTrip(goodRequest)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrStoreRead)
		mockRouter.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("should return a response if the store returns a snapshot", func(t *testing.T) {
		mockRouter, mockStore, player := newMockReplayTransport()
		mockRouter.On("Match", mock.Anything).Return(&spec.SnapshotRef{}, nil)
		mockStore.On("Read", mock.Anything).Return(&spec.Snapshot{
			Response: &spec.SerializableResponse{
				StatusCode: http.StatusOK,
				HTTPMessage: spec.HTTPMessage{
					Header: http.Header{},
				},
			},
		}, nil)
		res, err := player.RoundTrip(goodRequest)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, "true", res.Header.Get("X-Wiretap-Cache-Hit"))
		mockRouter.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})
}
