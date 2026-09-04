# Decorators

Format: `//kibu:<name>[:qualifier] key=value …`

Name every spec parameter and result: `ctx`, `req`, `res`, `err`. Queries omit `ctx`. Raw HTTP uses `tctx transport.Context`.

When this table and the Starlight docs disagree, follow testdata (`internal/toolchain/kibugenv2/testdata`) and existing `*.gen.go`.

| Decorator | Target | Notes |
|-----------|--------|--------|
| `//kibu:service` | interface | HTTP service. Optional `public`. |
| `//kibu:service:method` | method | `method=GET\|POST\|PUT\|DELETE\|PATCH` (default POST). `path=/…`. `mode=raw` → `transport.Context`, excluded from TS. |
| `//kibu:workflow` | interface | Temporal workflow. Optional `task_queue=…`. |
| `//kibu:workflow:execute` | method | `(ctx workflow.Context, req T) (res U, err error)` |
| `//kibu:workflow:signal` | method | `(ctx workflow.Context, req T) error` |
| `//kibu:workflow:query` | method | `(req T) (res U, err error)` — no context, no activities, no mutation. |
| `//kibu:workflow:update` | method | `(ctx workflow.Context, req T) (res U, err error)` — present in the generator; often missing from docs. |
| `//kibu:activity` | interface | Optional `task_queue=…`. |
| `//kibu:activity:method` | method | `(ctx context.Context, req T) (res U, err error)` |
| `//kibu:provider` | func/type/var | kibuwire. Optional `import=…` `group=…`. |

## Spec skeleton

```go
//kibu:service
type Service interface {
    //kibu:service:method method=GET path=/health
    Check(ctx context.Context, req CheckRequest) (res CheckResponse, err error)
}

//kibu:activity
type Activities interface {
    //kibu:activity:method
    ChargePaymentMethod(ctx context.Context, req ChargePaymentMethodRequest) (res ChargePaymentMethodResponse, err error)
}

//kibu:workflow
type CustomerSubscriptionsWorkflow interface {
    //kibu:workflow:execute
    Execute(ctx workflow.Context, req CustomerSubscriptionsRequest) (res CustomerSubscriptionsResponse, err error)

    //kibu:workflow:update
    AttemptPayment(ctx workflow.Context, req AttemptPaymentRequest) (res AttemptPaymentResponse, err error)

    //kibu:workflow:signal
    SetDiscount(ctx workflow.Context, req SetDiscountRequest) error

    //kibu:workflow:query
    GetAccountDetails(req GetAccountDetailsRequest) (res GetAccountDetailsResponse, err error)
}
```
