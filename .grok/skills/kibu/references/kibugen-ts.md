# kibugen_ts

```text
kibugen_ts [-cwd <dir>] [-out <dir>] <patterns…>
```

CLI flag is `-out` (docs sometimes say `-output`). Default `-out` is `<cwd>/gen`.

Emits TypeScript types and HTTP clients from `//kibu:service` methods. Skips `mode=raw`.

| Go | TypeScript |
|----|------------|
| `bool` | `boolean` |
| integer / float | `number` |
| `string` | `string` |
| `time.Time`, `uuid.UUID` | `string` |
| `[]T` | `T[]` |
| `map[K]V` | `Record<K, V>` |
| `*T` | `T \| null` optional field |

Change the Go spec and regenerate. Do not hand-edit generated TS to match a new field.

Requires kibumod discovery of `//kibu:service`. Workflows and activities are not TS clients.
