---
title: kibugenv2
description: Code generator for Kibu service plumbing including workflows, activities, and controllers
sidebar:
    order: 3
---

# kibugenv2 - Code Generation for System Plumbing

`kibugenv2` generates Go boilerplate code for Kibu application infrastructure. It creates type-safe interfaces, clients, and controllers for workflows, activities, and HTTP services.

## Installation

```shell
go install github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/cmd/kibugenv2@main
```

## Purpose

`kibugenv2` eliminates repetitive boilerplate by generating:

- **Workflow interfaces**: `WorkflowRun`, `WorkflowClient`, `WorkflowChildClient`
- **Activity interfaces**: `ActivityClient` for calling activities from workflows
- **Controllers**: Service, activity, workflow, and worker controllers
- **Signal channels**: For Temporal workflow signaling
- **Proxy implementations**: For calling workflows and activities
- **Compiler assertions**: Ensuring type safety at compile time

## Usage

Add a `go:generate` directive to your project:

```go
package generate

//go:generate kibugenv2 ./...
```

Then run:

```shell
go generate ./...
```

This produces `*.gen.go` files alongside your service definitions.

## Supported Decorators

### Service Decorators

| Decorator | Purpose | Options |
|-----------|---------|---------|
| `//kibu:service` | Define an HTTP service interface | - |
| `//kibu:service:method` | Define an HTTP endpoint | `method=GET\|POST\|PUT\|DELETE`, `path=/route` |

### Workflow Decorators

| Decorator | Purpose | Options |
|-----------|---------|---------|
| `//kibu:workflow` | Define a Temporal workflow interface | - |
| `//kibu:workflow:execute` | Mark the main workflow execution method | - |
| `//kibu:workflow:signal` | Define a signal handler | - |
| `//kibu:workflow:query` | Define a query handler | - |

### Activity Decorators

| Decorator | Purpose | Options |
|-----------|---------|---------|
| `//kibu:activity` | Define a Temporal activity interface | - |

## Generated Artifacts

### For Workflows

Given a workflow interface:

```go
//kibu:workflow
type OrderWorkflows interface {
    //kibu:workflow:execute
    ProcessOrder(ctx workflow.Context, req ProcessOrderRequest) (ProcessOrderResponse, error)

    //kibu:workflow:signal
    CancelOrder(ctx workflow.Context, req CancelOrderRequest) error

    //kibu:workflow:query
    GetStatus(ctx workflow.Context, req GetStatusRequest) (GetStatusResponse, error)
}
```

`kibugenv2` generates:

#### WorkflowRun Interface
```go
type ProcessOrderWorkflowRun interface {
    ID() string
    RunID() string
    Get(ctx context.Context) (ProcessOrderResponse, error)
    CancelOrder(ctx context.Context, req CancelOrderRequest) error
    GetStatus(ctx context.Context, req GetStatusRequest) (GetStatusResponse, error)
}
```

#### WorkflowClient Interface
```go
type ProcessOrderWorkflowClient interface {
    Execute(ctx context.Context, req ProcessOrderRequest, opts ...client.StartWorkflowOptions) (ProcessOrderWorkflowRun, error)
    GetHandle(ctx context.Context, workflowID, runID string) ProcessOrderWorkflowRun
}
```

#### WorkflowChildClient Interface
For calling workflows from other workflows:
```go
type ProcessOrderWorkflowChildClient interface {
    Execute(ctx workflow.Context, req ProcessOrderRequest, opts ...workflow.ChildWorkflowOptions) ProcessOrderWorkflowChildRun
}
```

### For Activities

Given an activity interface:

```go
//kibu:activity
type OrderActivities interface {
    ValidateOrder(ctx context.Context, req ValidateOrderRequest) (ValidateOrderResponse, error)
    ChargePayment(ctx context.Context, req ChargePaymentRequest) (ChargePaymentResponse, error)
}
```

`kibugenv2` generates:

#### ActivityClient Interface
```go
type OrderActivitiesClient interface {
    ValidateOrder(ctx workflow.Context, req ValidateOrderRequest) (ValidateOrderResponse, error)
    ChargePayment(ctx workflow.Context, req ChargePaymentRequest) (ChargePaymentResponse, error)
}
```

### For Services

Given a service interface:

```go
//kibu:service
type HealthService interface {
    //kibu:service:method method=GET path=/health
    Check(ctx context.Context, req CheckRequest) (CheckResponse, error)
}
```

`kibugenv2` generates:

#### ServiceController
```go
type HealthServiceController struct {
    impl HealthService
}

func (c *HealthServiceController) RegisterRoutes(r *gin.Engine) {
    r.GET("/health", c.handleCheck)
}
```

### Controllers

`kibugenv2` generates several controller types:

- **ServiceController**: HTTP route handlers
- **ActivityController**: Temporal activity registration
- **WorkflowController**: Temporal workflow registration
- **WorkerController**: Combined worker setup

## Example Project Structure

```
src/backend/systems/orders/
├── orders.go           # Your service definitions
├── orders.gen.go       # Generated plumbing (auto-generated)
├── orders_impl.go      # Your implementations
└── orders_test.go      # Tests
```

## Integration with Wire

Generated controllers include Wire provider annotations for dependency injection:

```go
//kibu:provider import=github.com/kibu-sh/kibu/pkg/transport/httpx group=HandlerFactory
type HealthServiceController struct { ... }
```

See [kibuwire](/reference/toolchain/kibuwire) for more on dependency injection integration.

## Best Practices

1. **Keep interfaces focused**: One workflow or service per interface
2. **Use meaningful names**: Interface names become part of generated type names
3. **Document your interfaces**: Comments are preserved in generated code
4. **Don't edit generated files**: Changes will be overwritten on regeneration

## Troubleshooting

### Generated files not updating

Ensure you run `go generate` after modifying your interfaces:

```shell
go generate ./...
```

### Type errors in generated code

Check that your interface methods follow the expected signature pattern:

- First parameter should be `context.Context` (services/activities) or `workflow.Context` (workflows)
- Return type should be `(ResponseType, error)`

### Missing imports

`kibugenv2` automatically handles imports. If you see import errors, ensure your Go module is properly initialized.
