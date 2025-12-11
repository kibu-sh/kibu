---
title: Toolchain Overview
description: Kibu's code generation and analysis toolchain
sidebar:
    order: 1
---

# Kibu Toolchain

Kibu provides a suite of code generation and analysis tools that eliminate boilerplate and enable type-safe communication across your stack.

## Core Tools

| Tool | Purpose | Documentation |
|------|---------|---------------|
| **kibumod** | Service definition analyzer | [Reference](/reference/toolchain/kibumod) |
| **kibugenv2** | Go plumbing code generator | [Reference](/reference/toolchain/kibugenv2) |
| **kibuwire** | Wire dependency injection generator | [Reference](/reference/toolchain/kibuwire) |
| **kibugen_ts** | TypeScript client generator | [Reference](/reference/toolchain/kibugen-ts) |

## How It Works

The toolchain follows a pipeline pattern:

```
Go Source Code
     │
     ▼
┌─────────┐
│ kibumod │  ← Analyzes interfaces with //kibu: decorators
└────┬────┘
     │
     ▼
┌───────────┐     ┌────────────┐
│ kibugenv2 │ ──► │ kibuwire   │
└─────┬─────┘     └─────┬──────┘
      │                 │
      ▼                 ▼
  *.gen.go        kibuwire.gen.go
      │
      ▼
┌────────────┐
│ kibugen_ts │
└─────┬──────┘
      │
      ▼
   *.gen.ts
```

## Quick Start

Install the toolchain:

```shell
go install github.com/google/wire/cmd/wire@latest
go install github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/cmd/kibugenv2@main
go install github.com/kibu-sh/kibu/internal/toolchain/kibuwire/cmd/kibuwire@main
go install github.com/kibu-sh/kibu/internal/toolchain/kibugen_ts/cmd/kibugen_ts@main
```

Create a `generate.go` file:

```go
package generate

//go:generate kibugenv2 ./...
//go:generate kibuwire ./...
//go:generate kibugen_ts -output ./frontend/src/api ./...
```

Run generation:

```shell
go generate ./...
```

## Decorator-Driven Development

Kibu uses comment directives (decorators) to annotate your code:

```go
// UserService handles user operations
//
//kibu:service
type UserService interface {
    //kibu:service:method method=GET path=/users/{id}
    GetUser(ctx context.Context, req GetUserRequest) (GetUserResponse, error)
}
```

See the [Decorators Reference](/reference/decorators) for all available decorators.

## Generated Artifacts

| Tool | Output | Purpose |
|------|--------|---------|
| kibugenv2 | `*.gen.go` | Controllers, clients, workflow interfaces |
| kibuwire | `kibuwire/kibuwire.gen.go` | Wire provider sets |
| kibugen_ts | `*.gen.ts` | TypeScript types and service clients |

## Best Practices

1. **Don't edit generated files** - They're overwritten on each generation
2. **Commit generated files** - Include in version control for CI/CD
3. **Run in order** - kibugenv2 before kibuwire, both before kibugen_ts
4. **Use go:generate** - Centralize generation in a single file