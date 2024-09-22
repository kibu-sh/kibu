package temporal

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

type SignalChannel[T any] interface {
	// Receive blocks until the signalChannel is received
	// more is false if the channel was closed
	Receive(ctx workflow.Context) (res T, more bool)

	// ReceiveAsync checks for a signalChannel without blocking
	// returns ok of false when no value was found in the channel
	ReceiveAsync() (res T, ok bool)

	// ReceiveWithTimeout blocks until a signalChannel is received or the timeout expires.
	// Returns more value of false when Channel is closed.
	// Returns ok value of false when no value was found in the channel for the duration of timeout or the ctx was canceled.
	ReceiveWithTimeout(ctx workflow.Context, timeout time.Duration) (res T, ok bool, more bool)

	// Select checks for signalChannel without blocking
	Select(sel workflow.Selector, fn func(res T, ok bool)) workflow.Selector

	// Len returns the number of elements in the channel
	Len() int
}

var _ SignalChannel[any] = (*signalChannel[any])(nil)

type signalChannel[T any] struct {
	channel workflow.ReceiveChannel
}

// Len returns the number of buffered messages plus the number of blocked Send calls.
func (s *signalChannel[T]) Len() int {
	return s.channel.Len()
}

// Receive blocks until the signalChannel is received
func (s *signalChannel[T]) Receive(ctx workflow.Context) (resp T, more bool) {
	more = s.channel.Receive(ctx, &resp)
	return
}

// ReceiveAsync checks for a signalChannel without blocking
// returns ok of false when no value was found in the channel
func (s *signalChannel[T]) ReceiveAsync() (res T, ok bool) {
	ok = s.channel.ReceiveAsync(&res)
	return
}

// ReceiveWithTimeout blocks until a signalChannel is received or the timeout expires.
// Returns more value of false when Channel is closed.
// Returns ok value of false when no value was found in the channel for the duration of timeout or the ctx was canceled.
// resp will be nil if ok is false.
func (s *signalChannel[T]) ReceiveWithTimeout(ctx workflow.Context, timeout time.Duration) (resp T, ok bool, more bool) {
	ok, more = s.channel.ReceiveWithTimeout(ctx, timeout, &resp)
	return
}

// Select registers a callback function to be called when a channel has a message to receive.
// The callback is called when Select(ctx) is called.
// The message is expected to be consumed by the callback function.
// The branch is automatically removed after the channel is closed, and the callback fires.
func (s *signalChannel[T]) Select(sel workflow.Selector, fn func(T, bool)) workflow.Selector {
	return sel.AddReceive(s.channel, func(workflow.ReceiveChannel, bool) {
		req, ok := s.ReceiveAsync()
		if fn != nil {
			fn(req, ok)
		}
	})
}

func NewSignalChannel[T any](ctx workflow.Context, signalName string) SignalChannel[T] {
	return &signalChannel[T]{workflow.GetSignalChannel(ctx, signalName)}
}
