# Implementations

Hand-written wiring is three `//kibu:provider` constructors in the system package. Generated controllers take `Service`, `Activities`, and `XxxWorkflowFactory` as fields. Generated `NewWorkflowsClient`, `NewActivitiesProxy`, and `*Controller` types come from `*.gen.go` / kibuwire — do not reimplement them.

The spec interface is satisfied by a private struct with receivers. Do not invent a second interface, and do not implement the spec with a func type (`type serviceFunc func(...)`). Keep method bodies on the receiver; do not extract a package function per method unless there is a real reuse or test need.

Name every parameter and result: `ctx`, `req`, `res`, `err` (and `input` / `wf` on factories). Do not write anonymous `(Foo, error)`.

Signatures for spec methods: [decorators.md](decorators.md).

## Shape

1. Exported `XxxDeps` with the collaborators Wire should inject.
2. Unexported impl struct whose only injected field is `deps XxxDeps`.
3. Provider takes `XxxDeps` and assigns it: `return &service{deps: deps}`.
4. Interface methods are receivers that use `s.deps`.

Mark both the deps struct and the constructor `//kibu:provider` so kibuwire emits `wire.Struct(new(XxxDeps), "*")` plus the constructor.

## Service

```go
//kibu:provider
type ServiceDeps struct {
    Workflows WorkflowsClient
}

type service struct {
    deps ServiceDeps
}

//kibu:provider
func NewService(deps ServiceDeps) Service {
    return &service{deps: deps}
}

func (s *service) WatchAccount(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error) {
    run, err := s.deps.Workflows.CustomerSubscriptionsWorkflow().Execute(ctx, CustomerSubscriptionsRequest{})
    if err != nil {
        return
    }
    details, err := run.GetAccountDetails(ctx, GetAccountDetailsRequest{})
    if err != nil {
        return
    }
    res.Status = details.Status
    return
}
```

## Activities

Same shape: `ActivitiesDeps`, `activities struct { deps ActivitiesDeps }`, `NewActivities(deps ActivitiesDeps) Activities`. Receivers on `*activities`. The spec already has `Activities` — do not declare another interface.

Call activities from workflows through the generated `ActivitiesProxy`, not by constructing the activity impl inside the workflow.

## Workflow

Mutable workflow data lives on a pointer to an unexported state struct, not as fields next to `deps` / `input`.

```go
//kibu:provider
type CustomerSubscriptionsWorkflowDeps struct {
    Activities ActivitiesProxy
}

type customerSubscriptionsWorkflowState struct {
    accountStatus AccountStatus
    discountCode  string
}

type customerSubscriptionsWorkflow struct {
    deps  CustomerSubscriptionsWorkflowDeps
    input *CustomerSubscriptionsWorkflowInput
    state *customerSubscriptionsWorkflowState
}

//kibu:provider
func NewCustomerSubscriptionsWorkflowFactory(deps CustomerSubscriptionsWorkflowDeps) CustomerSubscriptionsWorkflowFactory {
    return func(input *CustomerSubscriptionsWorkflowInput) (wf CustomerSubscriptionsWorkflow, err error) {
        wf = &customerSubscriptionsWorkflow{
            deps:  deps,
            input: input,
            state: &customerSubscriptionsWorkflowState{},
        }
        return
    }
}
```

Execute / signal / query / update are receivers on `*customerSubscriptionsWorkflow`. Read and write `s.state`, call activities via `s.deps`, receive signals from `s.input`. Queries must not call activities.

## Compile

After adding providers, run kibuwire then `wire ./...`. Missing-provider errors mean a constructor or deps struct is missing `//kibu:provider`, or the constructor return type is not the spec interface.
