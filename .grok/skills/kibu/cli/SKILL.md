---
name: kibu-cli
description: >
  This skill should be used when reading or writing the Kibu config store,
  secrets, encrypted .enc.json files, or when the user runs /kibu-cli, kibu
  config, kibu migrate, kibu dev, kibu build, or asks to inspect or change
  configuration with a temporary Go program.
---

# kibu CLI

Use the `kibu` binary for operator tasks. Do not create a temporary `main.go` that imports `pkg/config` or `pkg/workspace` to get or set values.

## Config (default)

Persistent `-e` / `--environment` (default `dev`). Keys are store paths, not filenames.

```bash
kibu config get <key> -e <env>
kibu config set <key> -e <env> --from-literal k=v
kibu config set <key> -e <env> --from-file data.json
kibu config set <key> -e <env> --from-env-file .env
kibu config edit <key> -e <env>          # $EDITOR; wait for the editor to exit
kibu config copy <path> --src <env> --dest <env> [-r]
kibu config sync -e <env> --google-project <id>
```

`.kibu/store/config/<env>/<key>.enc.json` is ciphertext. Read plaintext with `kibu config get`.

Application runtime: inject `config.Store` (for example `wireset.DefaultSet` → `GetByKey`). That is production code, not a session scratch program.

Store layout and encryption: [../references/config.md](../references/config.md).

## Other commands

`dev up`, `migrate up|down`, `build`: [../references/cli.md](../references/cli.md).
