package temporal

import (
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"time"

	"go.temporal.io/sdk/workflow"
)

type WorkflowOptionProvider interface {
	WorkflowOptions(builder WorkflowOptionsBuilder) WorkflowOptionsBuilder
}

type WorkflowOptionFunc func(WorkflowOptionsBuilder) WorkflowOptionsBuilder

// WorkflowOptionsBuilder helps build both StartWorkflowOptions and ChildWorkflowOptions.
type WorkflowOptionsBuilder struct {
	// Common fields
	id                                     string
	taskQueue                              string
	workflowExecutionTimeout               time.Duration
	workflowRunTimeout                     time.Duration
	workflowTaskTimeout                    time.Duration
	memo                                   map[string]any
	retryPolicy                            *temporal.RetryPolicy
	workflowIDReusePolicy                  enums.WorkflowIdReusePolicy
	versioningIntent                       temporal.VersioningIntent
	workflowIDConflictPolicy               enums.WorkflowIdConflictPolicy
	enableEagerStart                       bool
	startDelay                             time.Duration
	staticSummary                          string
	staticDetails                          string
	workflowExecutionErrorIfAlreadyStarted bool

	// StartWorkflowOptions specific
	cronSchedule       string
	withStartOperation client.WithStartWorkflowOperation

	// ChildWorkflowOptions specific
	namespace           string
	parentClosePolicy   enums.ParentClosePolicy
	waitForCancellation bool
}

// NewWorkflowOptionsBuilder creates a new WorkflowOptionsBuilder.
func NewWorkflowOptionsBuilder() WorkflowOptionsBuilder {
	return WorkflowOptionsBuilder{}
}

// WithID sets the ID for StartWorkflowOptions and WorkflowID for ChildWorkflowOptions.
func (b WorkflowOptionsBuilder) WithID(id string) WorkflowOptionsBuilder {
	b.id = id
	return b
}

// WithTaskQueue sets the task queue.
func (b WorkflowOptionsBuilder) WithTaskQueue(taskQueue string) WorkflowOptionsBuilder {
	b.taskQueue = taskQueue
	return b
}

// WithWorkflowExecutionTimeout sets the workflow execution timeout.
func (b WorkflowOptionsBuilder) WithWorkflowExecutionTimeout(d time.Duration) WorkflowOptionsBuilder {
	b.workflowExecutionTimeout = d
	return b
}

// WithWorkflowRunTimeout sets the workflow run timeout.
func (b WorkflowOptionsBuilder) WithWorkflowRunTimeout(d time.Duration) WorkflowOptionsBuilder {
	b.workflowRunTimeout = d
	return b
}

// WithWorkflowTaskTimeout sets the workflow task timeout.
func (b WorkflowOptionsBuilder) WithWorkflowTaskTimeout(d time.Duration) WorkflowOptionsBuilder {
	b.workflowTaskTimeout = d
	return b
}

// WithCronSchedule sets the cron schedule (for StartWorkflowOptions).
func (b WorkflowOptionsBuilder) WithCronSchedule(schedule string) WorkflowOptionsBuilder {
	b.cronSchedule = schedule
	return b
}

// WithMemo sets the memo.
func (b WorkflowOptionsBuilder) WithMemo(memo map[string]interface{}) WorkflowOptionsBuilder {
	b.memo = memo
	return b
}

// WithRetryPolicy sets the retry policy.
func (b WorkflowOptionsBuilder) WithRetryPolicy(policy *temporal.RetryPolicy) WorkflowOptionsBuilder {
	b.retryPolicy = policy
	return b
}

// WithWorkflowIDReusePolicy sets the workflow ID reuse policy.
func (b WorkflowOptionsBuilder) WithWorkflowIDReusePolicy(policy enums.WorkflowIdReusePolicy) WorkflowOptionsBuilder {
	b.workflowIDReusePolicy = policy
	return b
}

// WithVersioningIntent sets the versioning intent.
func (b WorkflowOptionsBuilder) WithVersioningIntent(intent temporal.VersioningIntent) WorkflowOptionsBuilder {
	b.versioningIntent = intent
	return b
}

// WithWorkflowIDConflictPolicy sets the workflow ID conflict policy.
func (b WorkflowOptionsBuilder) WithWorkflowIDConflictPolicy(policy enums.WorkflowIdConflictPolicy) WorkflowOptionsBuilder {
	b.workflowIDConflictPolicy = policy
	return b
}

// WithIdReusePolicy sets the workflow ID reuse policy.
func (b WorkflowOptionsBuilder) WithIdReusePolicy(policy enums.WorkflowIdReusePolicy) WorkflowOptionsBuilder {
	b.workflowIDReusePolicy = policy
	return b
}

// WithEnableEagerStart enables or disables eager workflow start.
func (b WorkflowOptionsBuilder) WithEnableEagerStart(enable bool) WorkflowOptionsBuilder {
	b.enableEagerStart = enable
	return b
}

// WithStartDelay sets the start delay duration.
func (b WorkflowOptionsBuilder) WithStartDelay(delay time.Duration) WorkflowOptionsBuilder {
	b.startDelay = delay
	return b
}

// WithStaticSummary sets the static summary.
func (b WorkflowOptionsBuilder) WithStaticSummary(summary string) WorkflowOptionsBuilder {
	b.staticSummary = summary
	return b
}

// WithStaticDetails sets the static details.
func (b WorkflowOptionsBuilder) WithStaticDetails(details string) WorkflowOptionsBuilder {
	b.staticDetails = details
	return b
}

// WithWorkflowExecutionErrorIfAlreadyStarted sets whether to return an error if the workflow is already started.
func (b WorkflowOptionsBuilder) WithWorkflowExecutionErrorIfAlreadyStarted(flag bool) WorkflowOptionsBuilder {
	b.workflowExecutionErrorIfAlreadyStarted = flag
	return b
}

// WithStartOperation sets the start operation option (for StartWorkflowOptions).
func (b WorkflowOptionsBuilder) WithStartOperation(option client.WithStartWorkflowOperation) WorkflowOptionsBuilder {
	b.withStartOperation = option
	return b
}

// WithNamespace sets the namespace (for ChildWorkflowOptions).
func (b WorkflowOptionsBuilder) WithNamespace(namespace string) WorkflowOptionsBuilder {
	b.namespace = namespace
	return b
}

// WithParentClosePolicy sets the parent close policy (for ChildWorkflowOptions).
func (b WorkflowOptionsBuilder) WithParentClosePolicy(policy enums.ParentClosePolicy) WorkflowOptionsBuilder {
	b.parentClosePolicy = policy
	return b
}

// WithWaitForCancellation sets the wait for cancellation flag (for ChildWorkflowOptions).
func (b WorkflowOptionsBuilder) WithWaitForCancellation(wait bool) WorkflowOptionsBuilder {
	b.waitForCancellation = wait
	return b
}

// ApplyOptionFuncs applies the provided option functions to the builder.
func (b WorkflowOptionsBuilder) ApplyOptionFuncs(funcs ...WorkflowOptionFunc) WorkflowOptionsBuilder {
	for _, f := range funcs {
		b = f(b)
	}
	return b
}

// ApplyOptionProviders applies the provided option providers to the builder.
func (b WorkflowOptionsBuilder) ApplyOptionProviders(providers ...WorkflowOptionProvider) WorkflowOptionsBuilder {
	for _, p := range providers {
		b = p.WorkflowOptions(b)
	}
	return b
}

// ApplyOptionProvidersWhenSupported applies the provided option providers to the builder if the builder supports them.
func (b WorkflowOptionsBuilder) ApplyOptionProvidersWhenSupported(providers ...any) WorkflowOptionsBuilder {
	for _, p := range providers {
		if supported, ok := p.(WorkflowOptionProvider); ok {
			b = supported.WorkflowOptions(b)
		}
	}
	return b
}

// AsStartOptions constructs the workflow.StartWorkflowOptions.
func (b WorkflowOptionsBuilder) AsStartOptions() client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		ID:                                       b.id,
		TaskQueue:                                b.taskQueue,
		WorkflowExecutionTimeout:                 b.workflowExecutionTimeout,
		WorkflowRunTimeout:                       b.workflowRunTimeout,
		WorkflowTaskTimeout:                      b.workflowTaskTimeout,
		CronSchedule:                             b.cronSchedule,
		Memo:                                     b.memo,
		RetryPolicy:                              b.retryPolicy,
		WorkflowIDReusePolicy:                    b.workflowIDReusePolicy,
		WorkflowIDConflictPolicy:                 b.workflowIDConflictPolicy,
		WorkflowExecutionErrorWhenAlreadyStarted: b.workflowExecutionErrorIfAlreadyStarted,
		EnableEagerStart:                         b.enableEagerStart,
		StartDelay:                               b.startDelay,
		StaticSummary:                            b.staticSummary,
		StaticDetails:                            b.staticDetails,
		WithStartOperation:                       b.withStartOperation,
	}
}

// AsChildOptions constructs the workflow.ChildWorkflowOptions.
func (b WorkflowOptionsBuilder) AsChildOptions() workflow.ChildWorkflowOptions {
	return workflow.ChildWorkflowOptions{
		Namespace:                b.namespace,
		WorkflowID:               b.id,
		TaskQueue:                b.taskQueue,
		WorkflowExecutionTimeout: b.workflowExecutionTimeout,
		WorkflowRunTimeout:       b.workflowRunTimeout,
		WorkflowTaskTimeout:      b.workflowTaskTimeout,
		WaitForCancellation:      b.waitForCancellation,
		WorkflowIDReusePolicy:    b.workflowIDReusePolicy,
		RetryPolicy:              b.retryPolicy,
		CronSchedule:             b.cronSchedule,
		Memo:                     b.memo,
		ParentClosePolicy:        b.parentClosePolicy,
		VersioningIntent:         b.versioningIntent,
	}
}
