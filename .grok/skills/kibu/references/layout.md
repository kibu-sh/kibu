# Layout

Do not prescribe how Postgres, Temporal, or other runtime dependencies are provisioned.

## App tree

```
generate.go
src/backend/
  cmd/server/          # main, wireinject, wire_set
  database/            # optional sqlc / migrations
  systems/<name>/      # one Go package per system
gen/kibuwire/          # kibuwire SuperSet when -out gen/
```

`generate.go`:

```go
package generate

//go:generate go run github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/cmd/kibugenv2 ./...
//go:generate go run github.com/kibu-sh/kibu/internal/toolchain/kibuwire/cmd/kibuwire -out gen/ ./...
//go:generate wire ./...
```

Match existing `generate.go` in the app if it already differs.

`cmd/server/wire_set.go` is the process graph. Include `kibuwire.SuperSet` once. Do not list each system.

```go
var wireSet = wire.NewSet(
    wireset.DefaultSet,
    kibuwire.SuperSet,
    NewWorkerOptions,
)
```

`InitServer()` is the wireinject entry in the same package.

## One package per system

Do not nest `services/`, `workflows/`, or `activities/` directories. That produces many identically named folders and extra import paths.

Split by files. Export only public interfaces and DTOs. Other systems import `.../systems/billingv1` and depend on `billingv1.Service`, not impl types.

```
src/backend/systems/billingv1/
  billingv1.spec.go     # exported interfaces + DTOs
  service.go            # NewService
  activities.go         # NewActivities
  workflows.go          # New…WorkflowFactory
  billingv1.gen.go      # generated — do not edit
```

All files are `package billingv1`.

Worked init example: sibling `kibu-sh/templates/starter`. Copy providers, factory, and server wire set. Do not copy its nested impl packages.

Implementations: [implementations.md](implementations.md).
