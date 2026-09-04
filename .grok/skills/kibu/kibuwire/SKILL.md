---
name: kibuwire
description: >
  This skill should be used when adding //kibu:provider constructors, generating
  Wire sets, debugging kibuwire.gen.go or SuperSet, or when the user runs
  /kibuwire or asks to run kibuwire / wire after kibugenv2.
---

# kibuwire

Collect `//kibu:provider` into Wire sets. Run after kibugenv2.

## Procedure

1. Confirm kibugenv2 has been run so generated controllers carry `//kibu:provider`.
2. Add hand-written providers (`NewService`, `NewActivities`, `New…WorkflowFactory`) in the system package. See [../references/implementations.md](../references/implementations.md).
3. Run `kibuwire` (typically `go run github.com/kibu-sh/kibu/internal/toolchain/kibuwire/cmd/kibuwire -out gen/ ./...`). Default `-out` is `gen/` under cwd.
4. Run `wire ./...` from packages with `wireinject` (usually `src/backend/cmd/server`).
5. Do not list each system in `wire_set.go`. Include `kibuwire.SuperSet` once.

Flags, groups, and output paths: [../references/kibuwire.md](../references/kibuwire.md).
