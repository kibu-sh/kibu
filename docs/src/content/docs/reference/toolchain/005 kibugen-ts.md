---
title: kibugen_ts
description: TypeScript client code generator for Kibu services
sidebar:
    order: 5
---

# kibugen_ts - TypeScript Client Generation

`kibugen_ts` generates TypeScript client code from Go service definitions, enabling type-safe communication between your Go backend and TypeScript/JavaScript frontends.

## Installation

```shell
go install github.com/kibu-sh/kibu/internal/toolchain/kibugen_ts/cmd/kibugen_ts@main
```

## Purpose

`kibugen_ts` bridges the gap between Go and TypeScript by:

1. Analyzing services discovered by `kibumod`
2. Converting Go types to TypeScript equivalents
3. Generating service factory functions that create HTTP clients
4. Handling complex type conversions including optionals, nested structs, and collections

## Usage

```shell
kibugen_ts -output ./src/api ./...
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-output` | Output directory for generated TypeScript files | `./generated` |

## Type Conversions

`kibugen_ts` automatically converts Go types to their TypeScript equivalents:

### Primitive Types

| Go Type | TypeScript Type |
|---------|-----------------|
| `bool` | `boolean` |
| `int`, `int8`, `int16`, `int32`, `int64` | `number` |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `number` |
| `float32`, `float64` | `number` |
| `string` | `string` |

### Special Types

| Go Type | TypeScript Type |
|---------|-----------------|
| `time.Time` | `string` |
| `uuid.UUID` | `string` |
| `uuid.NullUUID` | `string \| null` |

### Collections

| Go Type | TypeScript Type |
|---------|-----------------|
| `[]T` | `T[]` |
| `map[K]V` | `Record<K, V>` |

### Pointers and Optionals

Go pointers become optional nullable fields in TypeScript:

```go
// Go
type User struct {
    Name     string  `json:"name"`
    Nickname *string `json:"nickname"`
}
```

```typescript
// TypeScript
interface User {
    name: string;
    nickname?: string | null;
}
```

### Nested Structs

Nested struct types are fully qualified to avoid naming conflicts:

```go
// Go (package: backend/typesv1)
type User struct {
    Profile Profile `json:"profile"`
}

type Profile struct {
    Bio string `json:"bio"`
}
```

```typescript
// TypeScript
interface backend_typesv1_User {
    profile: backend_typesv1_Profile;
}

interface backend_typesv1_Profile {
    bio: string;
}
```

### Struct Embedding

Embedded struct fields are promoted to the parent type:

```go
// Go
type BaseModel struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"created_at"`
}

type User struct {
    BaseModel
    Name string `json:"name"`
}
```

```typescript
// TypeScript
interface User {
    id: string;
    created_at: string;
    name: string;
}
```

## Generated Output

### Type Definitions

For each Go struct used in service methods, a TypeScript interface is generated:

```typescript
export interface CheckRequest {
    name: string;
}

export interface CheckResponse {
    value: string;
    status: Status;
    status_list: Status[];
    optional_status?: Status | null;
}

export interface Status {
    code: string;
}
```

### Service Factory Functions

For each `//kibu:service` interface, a factory function is generated:

```go
// Go
//kibu:service
type HealthService interface {
    //kibu:service:method method=GET path=/health
    Check(ctx context.Context, req CheckRequest) (CheckResponse, error)
}
```

```typescript
// TypeScript
export function createHealthService(client: HttpClient) {
    return {
        async check(req: CheckRequest): Promise<CheckResponse> {
            return client.request({
                method: 'GET',
                path: '/health',
                body: req,
            });
        },
    };
}
```

### Namespace Exports

All service creators are grouped under a namespace for easy importing:

```typescript
export const Services = {
    createHealthServiceV1,
    createHealthServiceV2,
    createUserService,
    createOrderService,
};
```

## Example Usage

### In Your Frontend

```typescript
// api/client.ts
import { createHealthService } from './generated/health.gen';

const httpClient = {
    async request<T>(opts: { method: string; path: string; body?: any }): Promise<T> {
        const response = await fetch(`/api${opts.path}`, {
            method: opts.method,
            headers: { 'Content-Type': 'application/json' },
            body: opts.body ? JSON.stringify(opts.body) : undefined,
        });
        return response.json();
    },
};

export const healthService = createHealthService(httpClient);
```

```typescript
// components/HealthCheck.tsx
import { healthService } from '../api/client';

async function checkHealth() {
    const response = await healthService.check({ name: 'frontend' });
    console.log(response.status);
}
```

## Generated File Structure

```
generated/
├── health.gen.ts      # Health service types and client
├── users.gen.ts       # User service types and client
├── orders.gen.ts      # Order service types and client
└── client.ts          # HTTP client utilities
```

## JSON Tag Handling

Field names in TypeScript follow the `json` tag from Go structs:

```go
type User struct {
    FirstName string `json:"first_name"`  // becomes first_name in TS
    LastName  string `json:"last_name"`   // becomes last_name in TS
    Email     string `json:"email"`       // becomes email in TS
}
```

## Excluding Endpoints

Endpoints marked with `mode=raw` are excluded from TypeScript generation:

```go
//kibu:service:method method=GET mode=raw
StreamData(ctx context.Context, req StreamRequest) (StreamResponse, error)
```

## Integration with Build Pipeline

Add to your `generate.go`:

```go
package generate

//go:generate kibugenv2 ./...
//go:generate kibuwire ./...
//go:generate kibugen_ts -output ../frontend/src/api ./...
```

Or integrate with your frontend build:

```json
{
  "scripts": {
    "generate:api": "cd ../backend && go generate ./...",
    "prebuild": "npm run generate:api"
  }
}
```

## Best Practices

1. **Run after backend changes**: Regenerate TypeScript when Go service definitions change
2. **Commit generated files**: Include `*.gen.ts` in version control for frontend CI/CD
3. **Use consistent naming**: JSON tags create the contract between Go and TypeScript
4. **Handle nulls carefully**: Optional fields may be `undefined` or `null`

## Troubleshooting

### Missing types

Ensure all struct types used in service methods are exported (uppercase names).

### Incorrect field names

Check that `json` tags are present on struct fields. Without tags, Go field names are used as-is.

### Circular type references

`kibugen_ts` handles circular references by using type names. If you see issues, check for deeply nested circular structures.
