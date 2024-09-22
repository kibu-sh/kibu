package temporal

import (
	"go.temporal.io/sdk/temporal"
	"time"

	"go.temporal.io/sdk/workflow"
)

type ActivityOptionsProvider interface {
	ActivityOptions(builder ActivityOptionsBuilder) ActivityOptionsBuilder
}

type ActivityOptionFunc func(ActivityOptionsBuilder) ActivityOptionsBuilder

// ActivityOptionsBuilder helps build workflow.ActivityOptions.
type ActivityOptionsBuilder struct {
	taskQueue              string
	scheduleToCloseTimeout time.Duration
	scheduleToStartTimeout time.Duration
	startToCloseTimeout    time.Duration
	heartbeatTimeout       time.Duration
	waitForCancellation    bool
	activityID             string
	retryPolicy            *temporal.RetryPolicy
	disableEagerExecution  bool
	versioningIntent       temporal.VersioningIntent
}

// NewActivityOptionsBuilder creates a new ActivityOptionsBuilder.
func NewActivityOptionsBuilder() ActivityOptionsBuilder {
	return ActivityOptionsBuilder{}
}

// WithTaskQueue sets the task queue.
func (b ActivityOptionsBuilder) WithTaskQueue(taskQueue string) ActivityOptionsBuilder {
	b.taskQueue = taskQueue
	return b
}

func (b ActivityOptionsBuilder) WithProviders(providers ...ActivityOptionsProvider) ActivityOptionsBuilder {
	for _, p := range providers {
		b = p.ActivityOptions(b)
	}
	return b
}

func (b ActivityOptionsBuilder) WithProvidersWhenSupported(providers ...any) ActivityOptionsBuilder {
	for _, p := range providers {
		if p, ok := p.(ActivityOptionsProvider); ok {
			b = p.ActivityOptions(b)
		}
	}
	return b
}

// WithScheduleToCloseTimeout sets the ScheduleToCloseTimeout.
func (b ActivityOptionsBuilder) WithScheduleToCloseTimeout(d time.Duration) ActivityOptionsBuilder {
	b.scheduleToCloseTimeout = d
	return b
}

// WithScheduleToStartTimeout sets the ScheduleToStartTimeout.
func (b ActivityOptionsBuilder) WithScheduleToStartTimeout(d time.Duration) ActivityOptionsBuilder {
	b.scheduleToStartTimeout = d
	return b
}

// WithStartToCloseTimeout sets the StartToCloseTimeout.
func (b ActivityOptionsBuilder) WithStartToCloseTimeout(d time.Duration) ActivityOptionsBuilder {
	b.startToCloseTimeout = d
	return b
}

// WithHeartbeatTimeout sets the HeartbeatTimeout.
func (b ActivityOptionsBuilder) WithHeartbeatTimeout(d time.Duration) ActivityOptionsBuilder {
	b.heartbeatTimeout = d
	return b
}

// WithWaitForCancellation sets the WaitForCancellation flag.
func (b ActivityOptionsBuilder) WithWaitForCancellation(wait bool) ActivityOptionsBuilder {
	b.waitForCancellation = wait
	return b
}

// WithActivityID sets the ActivityID.
func (b ActivityOptionsBuilder) WithActivityID(activityID string) ActivityOptionsBuilder {
	b.activityID = activityID
	return b
}

// WithRetryPolicy sets the RetryPolicy.
func (b ActivityOptionsBuilder) WithRetryPolicy(policy *temporal.RetryPolicy) ActivityOptionsBuilder {
	b.retryPolicy = policy
	return b
}

// WithDisableEagerExecution sets the DisableEagerExecution flag.
func (b ActivityOptionsBuilder) WithDisableEagerExecution(disable bool) ActivityOptionsBuilder {
	b.disableEagerExecution = disable
	return b
}

// WithVersioningIntent sets the VersioningIntent.
func (b ActivityOptionsBuilder) WithVersioningIntent(intent temporal.VersioningIntent) ActivityOptionsBuilder {
	b.versioningIntent = intent
	return b
}

func (b ActivityOptionsBuilder) WithOptionFuncs(funcs ...ActivityOptionFunc) ActivityOptionsBuilder {
	for _, f := range funcs {
		b = f(b)
	}
	return b
}

// Build constructs the workflow.ActivityOptions.
func (b ActivityOptionsBuilder) Build() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		TaskQueue:              b.taskQueue,
		ScheduleToCloseTimeout: b.scheduleToCloseTimeout,
		ScheduleToStartTimeout: b.scheduleToStartTimeout,
		StartToCloseTimeout:    b.startToCloseTimeout,
		HeartbeatTimeout:       b.heartbeatTimeout,
		WaitForCancellation:    b.waitForCancellation,
		ActivityID:             b.activityID,
		RetryPolicy:            b.retryPolicy,
		DisableEagerExecution:  b.disableEagerExecution,
		VersioningIntent:       b.versioningIntent,
	}
}
