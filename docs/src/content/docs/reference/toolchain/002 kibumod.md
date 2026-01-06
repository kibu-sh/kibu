---
title: kibumod
description: Service definition analyzer for extracting Kibu service metadata from Go source code
sidebar:
    order: 2
---

# kibumod - Service Definition Analyzer

`kibumod` is a Go static analysis tool that scans Go source code to identify and extract Kibu service definitions. It uses the `golang.org/x/tools/go/analysis` framework to parse AST (Abstract Syntax Tree) and find interface types decorated with Kibu directives.

## Installation

```shell
go install github.com/kibu-sh/kibu/internal/toolchain/kibumod/cmd/kibumod@main
```

## Purpose

`kibumod` serves as the foundation for Kibu's code generation pipeline. It:

1. Scans Go packages for interfaces decorated with `//kibu:` comments
2. Extracts service metadata including name, documentation, and operations
3. Parses decorator options from comment directives
4. Returns a structured `modspecv2.Package` containing all discovered services

## Supported Decorators

`kibumod` recognizes interfaces decorated with the following directives:

| Decorator | Purpose |
|-----------|---------|
| `//kibu:service` | HTTP service interface |
| `//kibu:workflow` | Temporal workflow interface |
| `//kibu:activity` | Temporal activity interface |

## Usage Example

### Defining a Service

```go
package health

import "context"

type CheckRequest struct {
    Name string `json:"name"`
}

type CheckResponse struct {
    Status string `json:"status"`
}

// Service provides health check endpoints for the system
//
//kibu:service
type Service interface {
    // Check verifies the system is operational
    //
    //kibu:service:method method=GET
    Check(ctx context.Context, req CheckRequest) (res CheckResponse, err error)
}
```

### Defining a Workflow

```go
package orders

import (
    "go.temporal.io/sdk/workflow"
)

type ProcessOrderRequest struct {
    OrderID string `json:"order_id"`
}

type ProcessOrderResponse struct {
    Status string `json:"status"`
}

// Workflows handles order processing workflows
//
//kibu:workflow
type Workflows interface {
    // ProcessOrder handles the complete order fulfillment process
    //
    //kibu:workflow:execute
    ProcessOrder(ctx workflow.Context, req ProcessOrderRequest) (res ProcessOrderResponse, err error)
}
```

### Defining Activities

```go
package orders

import "context"

type ValidateOrderRequest struct {
    OrderID string `json:"order_id"`
}

type ValidateOrderResponse struct {
    Valid bool `json:"valid"`
}

// Activities contains activities for order processing
//
//kibu:activity
type Activities interface {
    // ValidateOrder checks if an order is valid
    ValidateOrder(ctx context.Context, req ValidateOrderRequest) (res ValidateOrderResponse, err error)
}
```

## Output Structure

`kibumod` produces a `modspecv2.Package` structure containing:

- **Name**: The Go package name
- **Services**: List of discovered service definitions, each containing:
  - **Name**: Interface name
  - **Doc**: Documentation from comments
  - **Decorators**: Parsed directive list
  - **Operations**: Methods defined on the interface
    - **Name**: Method name
    - **Doc**: Method documentation
    - **Params**: Input parameters with types
    - **Results**: Return types
    - **Decorators**: Method-level directives

## Integration with Other Tools

`kibumod` is typically not used directly. Instead, it serves as an analyzer dependency for:

- **[kibugenv2](/reference/toolchain/kibugenv2)**: Generates Go plumbing code from discovered services
- **[kibugen_ts](/reference/toolchain/kibugen-ts)**: Generates TypeScript client code from discovered services

## Decorator Syntax

Kibu decorators follow this general format:

```
//tool:name:qualifier option1=value1 option2=value2
```

- **tool**: Always `kibu` for Kibu directives
- **name**: The directive type (e.g., `service`, `workflow`, `activity`)
- **qualifier**: Optional sub-directive (e.g., `method`, `execute`, `signal`)
- **options**: Key-value pairs separated by `=`

### Examples

```go
//kibu:service
//kibu:service:method method=GET path=/health
//kibu:workflow
//kibu:workflow:execute
//kibu:workflow:signal
//kibu:workflow:query
//kibu:activity
```

See the [Decorators Reference](/reference/decorators) for a complete list of supported decorators.
