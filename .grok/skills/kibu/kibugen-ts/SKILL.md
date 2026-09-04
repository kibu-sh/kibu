---
name: kibugen-ts
description: >
  This skill should be used when generating TypeScript clients from Kibu HTTP
  services, or when the user runs /kibugen-ts, mentions kibugen_ts, or asks
  for generated frontend API clients from Go //kibu:service interfaces.
---

# kibugen_ts

Generate TypeScript clients from `//kibu:service` interfaces. Run after Go generation is stable.

## Procedure

1. Confirm HTTP services exist (`//kibu:service` + `//kibu:service:method`). Skip `mode=raw` endpoints; they are not emitted.
2. Run `kibugen_ts -out <dir> ./...` (CLI flag is `-out`, not `-output`). Default `-out` is `gen/` under cwd.
3. Do not invent TS types that already exist in the generated client. Change the Go spec and regenerate.

Details: [../references/kibugen-ts.md](../references/kibugen-ts.md).
