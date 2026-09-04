# kibu CLI

Persistent flags on `kibu`: `-e` / `--environment` (default `dev`), `--debug`.

| Command | Use |
|---------|-----|
| `kibu config get <key> -e <env>` | Decrypt and print JSON |
| `kibu config set <key> -e <env> --from-literal k=v` | Write from literals (repeat `--from-literal`) |
| `kibu config set <key> -e <env> --from-file <json>` | Write JSON (`-` = stdin) |
| `kibu config set <key> -e <env> --from-env-file <dotenv>` | Write a dotenv file |
| `kibu config edit <key> -e <env>` | Open `$EDITOR`, encrypt on save |
| `kibu config copy <path> --src <env> --dest <env> [-r]` | Copy (and re-encrypt) between envs |
| `kibu config sync -e <env> --google-project <id>` | Push local env files to GCP Secret Manager |
| `kibu dev up` | Watch the repo and rebuild/restart `src/backend/cmd/server` |
| `kibu migrate up --dir <url> --database-url <dsn>` | Apply migrations |
| `kibu migrate down --dir <url> --database-url <dsn>` | Roll back migrations |
| `kibu build [patterns…]` | Run the codegen pipeline into the workspace codegen output dir |

Config store details: [config.md](config.md).
