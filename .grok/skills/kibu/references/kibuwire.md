# kibuwire

```text
kibuwire [-cwd <dir>] [-out <dir>] <patterns…>
```

`-out` defaults to `<cwd>/gen`. Typical: `-out gen/` → `gen/kibuwire/kibuwire.gen.go` containing `SuperSet`.

Each package with providers also gets a `WireSet`. `SuperSet` is `wire.NewSet` of those package `WireSet`s. `cmd/server` includes `kibuwire.SuperSet` once.

## Provider decorator

```go
//kibu:provider
func NewService(deps ServiceDeps) Service { … }
```

`//kibu:provider` on a struct type emits `wire.Struct(new(T), "*")` (used for `XxxDeps` and generated controllers).

| Option | Meaning |
|--------|---------|
| `import=path` | Interface package for grouped providers |
| `group=Name` | Aggregation group |

Generated controllers already set groups:

| Group | Import | Used for |
|-------|--------|----------|
| `HandlerFactory` | `github.com/kibu-sh/kibu/pkg/transport/httpx` | HTTP `ServiceController` |
| `WorkerFactory` | `github.com/kibu-sh/kibu/pkg/transport/temporal` | `WorkerController` |

Hand-written `NewService` / `NewActivities` / `New…WorkflowFactory` usually need no `group`.

Run after kibugenv2, then `wire ./...`.
