package devx

//#StartWorkflowOptions: {
// // ID - The business identifier of the workflow execution.
// // Optional: defaulted to a uuid.
// ID: string
//
// // TaskQueue - The workflow tasks of the workflow are scheduled on the queue with this name.
// // This is also the name of the activity task queue on which activities are scheduled.
// // The workflow author can choose to override this using activity options.
// // Mandatory: No default.
// TaskQueue: string
//
// // WorkflowExecutionTimeout - The timeout for duration of workflow execution.
// // It includes retries and continue as new. Use WorkflowRunTimeout to limit execution time
// // of a single workflow run.
// // The resolution is seconds.
// // Optional: defaulted to unlimited.
// WorkflowExecutionTimeout: time.#Duration
//
// // WorkflowRunTimeout - The timeout for duration of a single workflow run.
// // The resolution is seconds.
// // Optional: defaulted to WorkflowExecutionTimeout.
// WorkflowRunTimeout: time.#Duration
//
// // WorkflowTaskTimeout - The timeout for processing workflow task from the time the worker
// // pulled this task. If a workflow task is lost, it is retried after this timeout.
// // The resolution is seconds.
// // Optional: defaulted to 10 secs.
// WorkflowTaskTimeout: time.#Duration
//
// // WorkflowIDReusePolicy - Whether server allow reuse of workflow ID, can be useful
// // for dedupe logic if set to RejectDuplicate.
// // Optional: defaulted to AllowDuplicate.
// WorkflowIDReusePolicy: enumspb.#WorkflowIdReusePolicy
//
// // When WorkflowExecutionErrorWhenAlreadyStarted is true, Client.ExecuteWorkflow will return an error if the
// // workflow id has already been used and WorkflowIDReusePolicy would disallow a re-run. If it is set to false,
// // rather than erroring a WorkflowRun instance representing the current or last run will be returned.
// //
// // Optional: defaults to false
// WorkflowExecutionErrorWhenAlreadyStarted: bool
//
// // RetryPolicy - Optional retry policy for workflow. If a retry policy is specified, in case of workflow failure
// // server will start new workflow execution if needed based on the retry policy.
// RetryPolicy?: null | #RetryPolicy @go(,*RetryPolicy)
//
// // CronSchedule - Optional cron schedule for workflow. If a cron schedule is specified, the workflow will run
// // as a cron based on the schedule. The scheduling will be based on UTC time. Schedule for next run only happen
// // after the current run is completed/failed/timeout. If a RetryPolicy is also supplied, and the workflow failed
// // or timeout, the workflow will be retried based on the retry policy. While the workflow is retrying, it won't
// // schedule its next run. If next schedule is due while workflow is running (or retrying), then it will skip that
// // schedule. Cron workflow will not stop until it is terminated or canceled (by returning temporal.CanceledError).
// // The cron spec is as following:
// // ┌───────────── minute (0 - 59)
// // │ ┌───────────── hour (0 - 23)
// // │ │ ┌───────────── day of the month (1 - 31)
// // │ │ │ ┌───────────── month (1 - 12)
// // │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
// // │ │ │ │ │
// // │ │ │ │ │
// // * * * * *
// CronSchedule: string
//
// // Memo - Optional non-indexed info that will be shown in list workflow.
// Memo: {...} @go(,map[string]interface{})
//
// // SearchAttributes - Optional indexed info that can be used in query of List/Scan/Count workflow APIs (only
// // supported when Temporal server is using ElasticSearch). The key and value type must be registered on Temporal server side.
// // Use GetSearchAttributes API to get valid key and corresponding value type.
// SearchAttributes: {...} @go(,map[string]interface{})
//}

#Workflow: {
	Name: string

	#Handler
}
