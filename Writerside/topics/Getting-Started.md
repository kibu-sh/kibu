# Getting Started

## Install
Install the cli and library by running the following commands in your terminal. 
```bash
go get github.com/discernhq/devx
go install github.com/discernhq/devx/cmd/devx@main
go install github.com/google/wire/cmd/wire@latest
devx init # does not work right now
# should copy .devx/ and cue.mod to your project
# wire generate is required to produce the first version with a generate directive
```

### devx.gen.go
```go
package discern

//go:generate go run -mod=readonly github.com/discernhq/devx/cmd/devx build ./src/backend/systems/...
```

## Recommended Structure
```
mkdir -p .devx/store
mkdir -p src/backend/cmd/server
mkdir -p src/backend/systems/health
mkdir -p src/backend/systems/health/healthspec
touch docker-compose.yml
touch .devx/workspace.cue
touch src/backend/cmd/server/server.go
touch src/backend/cmd/server/foreman.go
touch src/backend/cmd/server/wire.go
touch src/backend/cmd/server/wire_set.go
touch src/backend/systems/health/health.go
touch src/backend/systems/health/healthspec/health_dto.go
```

Once the wire.go and wire_set.go files have been created cd into the server directory and run wire once.