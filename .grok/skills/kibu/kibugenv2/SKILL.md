---
name: kibugenv2
description: >
  This skill should be used when adding or changing a Kibu service, workflow,
  or activity interface, generating Go plumbing, debugging *.gen.go, or when
  the user runs /kibugenv2 or asks to run kibugenv2.
---

# kibugenv2

Generate Go plumbing from decorated interfaces in the same package. Do not edit `*.gen.go`.

## Procedure

1. Read [../references/layout.md](../references/layout.md). Create or use one package under `src/backend/systems/<name>/`.
2. Write DTOs and decorated interfaces in `<name>.spec.go`. Read [../references/decorators.md](../references/decorators.md). Name every param and result (`ctx`, `req`, `res`, `err`).
3. Run generation from the module `generate.go` (or `kibugenv2 ./...` / `go run github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/cmd/kibugenv2 ./...`).
4. Implement providers in sibling files in the same package. Read [../references/implementations.md](../references/implementations.md). Functional constructors, not pointer-receiver structs.
5. If `//kibu:provider` appeared on generated controllers, continue with [../kibuwire/SKILL.md](../kibuwire/SKILL.md).
6. Compile. If generated types look wrong, fix the spec and regenerate — do not patch the gen file.

## When generation is a no-op

kibugenv2 skips packages with no `//kibu:service`, `//kibu:workflow`, or `//kibu:activity` interfaces.

## Canonical examples

- Spec and artifacts: `internal/toolchain/kibugenv2/testdata/scripts/000000_init.txtar` (this repo).
- Init (providers, factory, server wire set): sibling `kibu-sh/templates/starter`, flattened to one package per [../references/layout.md](../references/layout.md).

Generated type names: [../references/kibugenv2.md](../references/kibugenv2.md). Analyzer role: [../references/kibumod.md](../references/kibumod.md).
