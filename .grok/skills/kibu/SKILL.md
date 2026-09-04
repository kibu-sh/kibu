---
name: kibu
description: >
  This skill should be used when working in a Kibu app: adding or changing
  systems, //kibu: services, Temporal workflows or activities, generating
  plumbing, Wire providers, or TypeScript clients. Use when the user runs
  /kibu, mentions kibugenv2, kibuwire, kibugen_ts, or asks to generate Go
  plumbing from decorated interfaces.
---

# Kibu

Kibu generates HTTP and Temporal plumbing from decorated Go interfaces. Hand-write the spec and providers. Do not edit `*.gen.go`.

## Pipeline

```
//kibu: interfaces  →  kibugenv2 (*.gen.go)  →  kibuwire (SuperSet)  →  wire
                                              ↘ kibugen_ts (optional)
```

Order: kibugenv2, then kibuwire, then `wire ./...`. Run TypeScript generation only after Go generation is stable.

When docs and the generator disagree, follow testdata and `*.gen.go` in this repo: `internal/toolchain/kibugenv2/testdata`.

## Layout (short)

One Go package per system. Files, not `services/` / `workflows/` / `activities/` directories. Export interfaces and DTOs only. Do not prescribe how Postgres, Temporal, or other runtime deps are provisioned.

## Named results

Every spec method, func type, and implementation names parameters and results: `ctx`, `req`, `res`, `err` (plus `input` / `wf` on workflow factories). Do not write anonymous `(Foo, error)`.

## Functional providers

Implement with functions that close over deps, not pointer-receiver structs. Hand-written wiring is `NewService`, `NewActivities`, and `New…WorkflowFactory` with `//kibu:provider`.

## Read next

| Situation | Load |
|-----------|------|
| New app or where code goes | [references/layout.md](references/layout.md) |
| Implement or initialize a system | [references/implementations.md](references/implementations.md) |
| `//kibu:` syntax and signatures | [references/decorators.md](references/decorators.md) |
| Generate Go plumbing | [kibugenv2/SKILL.md](kibugenv2/SKILL.md), then [references/kibugenv2.md](references/kibugenv2.md) |
| How discovery works | [references/kibumod.md](references/kibumod.md) |
| Wire / `//kibu:provider` | [kibuwire/SKILL.md](kibuwire/SKILL.md) → [references/kibuwire.md](references/kibuwire.md) |
| TypeScript clients | [kibugen-ts/SKILL.md](kibugen-ts/SKILL.md) → [references/kibugen-ts.md](references/kibugen-ts.md) |
