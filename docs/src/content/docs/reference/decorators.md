---
title: Decorators Reference
description: Complete reference for all Kibu decorator directives
sidebar:
    order: 1
---

# Kibu Decorators Reference

Kibu uses comment-based decorators to annotate Go code with metadata that drives code generation. This reference documents all available decorators.

## Decorator Syntax

Kibu decorators follow the Go directive comment format:

```
//tool:name:qualifier option1=value1 option2=value2
```

| Part | Description | Required |
|------|-------------|----------|
| `tool` | Always `kibu` for Kibu directives | Yes |
| `name` | The directive type (e.g., `service`, `workflow`) | Yes |
| `qualifier` | Sub-directive for method-level annotations | No |
| `options` | Key-value pairs for configuration | No |

## Service Decorators

Used for defining HTTP service interfaces.

### //kibu:service

Marks an interface as an HTTP service.

```go
// UserService handles user management
//
//kibu:service
type UserService interface {
    // ...
}
```

**Used by**: `kibumod`, `kibugenv2`, `kibugen_ts`

### //kibu:service:method

Defines an HTTP endpoint on a service method.

```go
//kibu:service:method method=GET path=/users/{id}
GetUser(ctx context.Context, req GetUserRequest) (GetUserResponse, error)
```

| Option | Description | Values | Default |
|--------|-------------|--------|---------|
| `method` | HTTP method | `GET`, `POST`, `PUT`, `DELETE`, `PATCH` | `POST` |
| `path` | URL path pattern | Any valid path | Method name |
| `mode` | Endpoint mode | `raw` (excludes from TS generation) | - |

**Used by**: `kibugenv2`, `kibugen_ts`

## Workflow Decorators

Used for defining Temporal workflow interfaces.

### //kibu:workflow

Marks an interface as a Temporal workflow.

```go
// OrderWorkflows handles order processing workflows
//
//kibu:workflow
type OrderWorkflows interface {
    // ...
}
```

**Used by**: `kibumod`, `kibugenv2`

### //kibu:workflow:execute

Marks the main execution method of a workflow.

```go
//kibu:workflow:execute
ProcessOrder(ctx workflow.Context, req ProcessOrderRequest) (ProcessOrderResponse, error)
```

**Used by**: `kibugenv2`

### //kibu:workflow:signal

Defines a signal handler for workflow communication.

```go
//kibu:workflow:signal
CancelOrder(ctx workflow.Context, req CancelOrderRequest) error
```

**Used by**: `kibugenv2`

### //kibu:workflow:query

Defines a query handler for workflow state inspection.

```go
//kibu:workflow:query
GetStatus(ctx workflow.Context, req GetStatusRequest) (GetStatusResponse, error)
```

**Used by**: `kibugenv2`

## Activity Decorators

Used for defining Temporal activity interfaces.

### //kibu:activity

Marks an interface as a Temporal activity container.

```go
// OrderActivities contains activities for order processing
//
//kibu:activity
type OrderActivities interface {
    ValidateOrder(ctx context.Context, req ValidateOrderRequest) (ValidateOrderResponse, error)
    ChargePayment(ctx context.Context, req ChargePaymentRequest) (ChargePaymentResponse, error)
}
```

**Used by**: `kibumod`, `kibugenv2`

## Provider Decorators

Used for dependency injection with Google Wire.

### //kibu:provider

Marks a type, function, or variable as a Wire provider.

```go
//kibu:provider
func NewDatabase(cfg *Config) (*sql.DB, error) {
    return sql.Open("postgres", cfg.DSN)
}
```

| Option | Description | Example |
|--------|-------------|---------|
| `import` | Package path for the provider interface | `import=github.com/kibu-sh/kibu/pkg/transport/httpx` |
| `group` | Provider group for aggregation | `group=HandlerFactory` |

**Used by**: `kibuwire`

#### Provider with Group

```go
//kibu:provider import=github.com/kibu-sh/kibu/pkg/transport/httpx group=HandlerFactory
type ServiceController struct {
    impl Service
}
```

#### Provider Variable

```go
//kibu:provider
var WireSet = wire.NewSet(
    NewService,
    wire.Struct(new(ServiceController), "*"),
)
```

## Decorator Quick Reference

| Decorator | Target | Tool | Purpose |
|-----------|--------|------|---------|
| `//kibu:service` | Interface | kibumod, kibugenv2, kibugen_ts | HTTP service |
| `//kibu:service:method` | Method | kibugenv2, kibugen_ts | HTTP endpoint |
| `//kibu:workflow` | Interface | kibumod, kibugenv2 | Temporal workflow |
| `//kibu:workflow:execute` | Method | kibugenv2 | Workflow execution |
| `//kibu:workflow:signal` | Method | kibugenv2 | Workflow signal |
| `//kibu:workflow:query` | Method | kibugenv2 | Workflow query |
| `//kibu:activity` | Interface | kibumod, kibugenv2 | Temporal activity |
| `//kibu:provider` | Type/Func/Var | kibuwire | Wire provider |

## Complete Example

```go
package orders

import (
    "context"
    "go.temporal.io/sdk/workflow"
)

// Request/Response types
type CreateOrderRequest struct {
    CustomerID string `json:"customer_id"`
    Items      []Item `json:"items"`
}

type CreateOrderResponse struct {
    OrderID string `json:"order_id"`
    Status  string `json:"status"`
}

type ProcessOrderRequest struct {
    OrderID string `json:"order_id"`
}

type ProcessOrderResponse struct {
    Status string `json:"status"`
}

type CancelRequest struct {
    Reason string `json:"reason"`
}

type StatusRequest struct{}

type StatusResponse struct {
    Status string `json:"status"`
    Step   string `json:"step"`
}

// HTTP Service
//
//kibu:service
type Service interface {
    //kibu:service:method method=POST path=/orders
    CreateOrder(ctx context.Context, req CreateOrderRequest) (CreateOrderResponse, error)

    //kibu:service:method method=GET path=/orders/{id}
    GetOrder(ctx context.Context, req GetOrderRequest) (GetOrderResponse, error)
}

// Temporal Workflow
//
//kibu:workflow
type Workflows interface {
    //kibu:workflow:execute
    ProcessOrder(ctx workflow.Context, req ProcessOrderRequest) (ProcessOrderResponse, error)

    //kibu:workflow:signal
    Cancel(ctx workflow.Context, req CancelRequest) error

    //kibu:workflow:query
    GetStatus(ctx workflow.Context, req StatusRequest) (StatusResponse, error)
}

// Temporal Activities
//
//kibu:activity
type Activities interface {
    ValidateOrder(ctx context.Context, req ValidateOrderRequest) (ValidateOrderResponse, error)
    ReserveInventory(ctx context.Context, req ReserveInventoryRequest) (ReserveInventoryResponse, error)
    ChargePayment(ctx context.Context, req ChargePaymentRequest) (ChargePaymentResponse, error)
    ShipOrder(ctx context.Context, req ShipOrderRequest) (ShipOrderResponse, error)
}
```

## See Also

- [kibumod](/reference/toolchain/kibumod) - Service definition analyzer
- [kibugenv2](/reference/toolchain/kibugenv2) - Go code generation
- [kibuwire](/reference/toolchain/kibuwire) - Wire set generation
- [kibugen_ts](/reference/toolchain/kibugen-ts) - TypeScript generation
