package devx

//import (
// "time"
//)

//#RetryPolicy: {
// // Backoff interval for the first retry. If BackoffCoefficient is 1.0 then it is used for all retries.
// // If not set or set to 0, a default interval of 1s will be used.
// InitialInterval: time.#Duration
//
// // Coefficient used to calculate the next retry backoff interval.
// // The next retry interval is previous interval multiplied by this coefficient.
// // Must be 1 or larger. Default is 2.0.
// BackoffCoefficient: float64
//
// // Maximum backoff interval between retries. Exponential backoff leads to interval increase.
// // This value is the cap of the interval. Default is 100x of initial interval.
// MaximumInterval: time.#Duration
//
// // Maximum number of attempts. When exceeded the retries stop even if not expired yet.
// // If not set or set to 0, it means unlimited, and rely on activity ScheduleToCloseTimeout to stop.
// MaximumAttempts: int32
//
// // Non-Retriable errors. This is optional. Temporal server will stop retry if error type matches this list.
// // Note:
// //  - cancellation is not a failure, so it won't be retried,
// //  - only StartToClose or Heartbeat timeouts are retryable.
// NonRetryableErrorTypes: [...string]
//}
//
//#ActivityOptions: {
// // TaskQueue that the activity needs to be scheduled on.
// // optional: The default task queue with the same name as the workflow task queue.
// TaskQueue: string
//
// // ScheduleToCloseTimeout - Total time that a workflow is willing to wait for Activity to complete.
// // ScheduleToCloseTimeout limits the total time of an Activity's execution including retries
// //   (use StartToCloseTimeout to limit the time of a single attempt).
// // The zero value of this uses default value.
// // Either this option or StartToClose is required: Defaults to unlimited.
// ScheduleToCloseTimeout: time.#Duration
//
// // ScheduleToStartTimeout - Time that the Activity Task can stay in the Task Queue before it is picked up by
// // a Worker. Do not specify this timeout unless using host specific Task Queues for Activity Tasks are being
// // used for routing. In almost all situations that don't involve routing activities to specific hosts it is
// // better to rely on the default value.
// // ScheduleToStartTimeout is always non-retryable. Retrying after this timeout doesn't make sense as it would
// // just put the Activity Task back into the same Task Queue.
// // If ScheduleToClose is not provided then this timeout is required.
// // Optional: Defaults to unlimited.
// ScheduleToStartTimeout: time.#Duration
//
// // StartToCloseTimeout - Maximum time of a single Activity execution attempt.
// // Note that the Temporal Server doesn't detect Worker process failures directly. It relies on this timeout
// // to detect that an Activity that didn't complete on time. So this timeout should be as short as the longest
// // possible execution of the Activity body. Potentially long running Activities must specify HeartbeatTimeout
// // and call Activity.RecordHeartbeat(ctx, "my-heartbeat") periodically for timely failure detection.
// // If ScheduleToClose is not provided then this timeout is required: Defaults to the ScheduleToCloseTimeout value.
// StartToCloseTimeout: time.#Duration
//
// // HeartbeatTimeout - Heartbeat interval. Activity must call Activity.RecordHeartbeat(ctx, "my-heartbeat")
// // before this interval passes after the last heartbeat or the Activity starts.
// HeartbeatTimeout: time.#Duration
//
// // WaitForCancellation - Whether to wait for canceled activity to be completed(
// // activity can be failed, completed, cancel accepted)
// // Optional: default false
// WaitForCancellation: bool
//
// // ActivityID - Business level activity ID, this is not needed for most of the cases if you have
// // to specify this then talk to temporal team. This is something will be done in future.
// // Optional: default empty string
// ActivityID: string
//
// // RetryPolicy specifies how to retry an Activity if an error occurs.
// // More details are available at docs.temporal.io.
// // RetryPolicy is optional. If one is not specified a default RetryPolicy is provided by the server.
// // The default RetryPolicy provided by the server specifies:
// // - InitialInterval of 1 second
// // - BackoffCoefficient of 2.0
// // - MaximumInterval of 100 x InitialInterval
// // - MaximumAttempts of 0 (unlimited)
// // To disable retries set MaximumAttempts to 1.
// // The default RetryPolicy provided by the server can be overridden by the dynamic config.
// RetryPolicy?: null | #RetryPolicy @go(,*RetryPolicy)
//
// // If true, will not request eager execution regardless of worker settings.
// // If false, eager execution may still be disabled at the worker level or
// // eager execution may not be requested due to lack of available slots.
// //
// // Eager activity execution means the server returns requested eager
// // activities directly from the workflow task back to this worker which is
// // faster than non-eager which may be dispatched to a separate worker.
// DisableEagerExecution: bool
//}

#Activity: {
	Name: string
	#Handler
	// Options: #ActivityOptions
}
