# Implementations

Hand-written wiring is three `//kibu:provider` constructors in the system package. Generated controllers take `Service`, `Activities`, and `XxxWorkflowFactory` as fields. Generated `NewWorkflowsClient`, `NewActivitiesProxy`, and `*Controller` types come from `*.gen.go` / kibuwire — do not reimplement them.

Prefer functions that close over deps. Avoid `type service struct { … }; func (s *service) Method(…)`.

Name every parameter and result: `ctx`, `req`, `res`, `err` (and `input` / `wf` on factories). Do not write anonymous `(Foo, error)`. Early `return` after `if err != nil` is the point of named results.

Signatures for spec methods: [decorators.md](decorators.md).

## Service

Single method → func type that implements the interface:

```go
type serviceFunc func(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error)

func (f serviceFunc) WatchAccount(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error) {
    return f(ctx, req)
}

//kibu:provider
func NewService(workflows WorkflowsClient) Service {
    return serviceFunc(func(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error) {
        run, err := workflows.CustomerSubscriptionsWorkflow().Execute(ctx, CustomerSubscriptionsRequest{})
        if err != nil {
            return
        }
        details, err := run.GetAccountDetails(ctx, GetAccountDetailsRequest{})
        if err != nil {
            return
        }
        res.Status = details.Status
        return
    })
}
```

Multi-method: a struct of funcs (one field per method), still constructed in `NewService`. Do not add pointer-receiver methods that read struct fields.

## Activities

Same pattern as Service. `NewActivities(deps) Activities` with `//kibu:provider`. Every activity method uses `(res …, err error)`.

Call activities from workflows through the generated `ActivitiesProxy`, not by constructing the activity impl inside the workflow.

## Workflow

Generated factory type (names may be un-named in `*.gen.go`; hand-written providers still name results):

```go
type CustomerSubscriptionsWorkflowFactory func(input *CustomerSubscriptionsWorkflowInput) (wf CustomerSubscriptionsWorkflow, err error)
```

```go
//kibu:provider
func NewCustomerSubscriptionsWorkflowFactory(activities ActivitiesProxy) CustomerSubscriptionsWorkflowFactory {
    return func(input *CustomerSubscriptionsWorkflowInput) (wf CustomerSubscriptionsWorkflow, err error) {
        wf = newCustomerSubscriptionsWorkflow(input, activities)
        return
    }
}
```

Workflows may keep a tiny state value because signals, queries, and updates share data. Construct it inside the factory. Do not hang workflow methods off a long-lived pointer service object.

Receive signals from `input.*Channel` or `New…SignalChannel(ctx)` as generated. Queries must not call activities.

## Compile

After adding providers, run kibuwire then `wire ./...`. Missing-provider errors mean a constructor is missing `//kibu:provider` or returns the wrong interface.
