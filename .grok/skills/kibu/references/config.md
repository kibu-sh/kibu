# Config store

Encrypted JSON objects, keyed by environment. Treat every key as sensitive.

Commands: [cli.md](cli.md).

On disk: `.kibu/store/config/<env>/<key>.enc.json`. Ciphertext (`EncryptionKey`, `Data`, `Version`) — not the secret. `kibu config get` decrypts.

CLI get/set/edit join `-e` with the key (`-e dev temporal` → `.kibu/store/config/dev/temporal.enc.json`). Default `-e` is `dev`.

`edit` waits on `$EDITOR`. Prefer `get` / `set` when the session cannot drive an editor.

`copy` re-encrypts with the `--dest` env key. `sync` uploads that env to GCP Secret Manager.

## Application code

Inject `config.Store`. `wireset.NewConfigStore` uses `workspace.DefaultConfigStore("dev")`, already scoped to that env, so keys are bare (`temporal`, not `dev/temporal`):

```go
_, err = store.GetByKey(ctx, "temporal", &opts)
```

Session operator work still uses `kibu config get`, not a scratch program that calls this.