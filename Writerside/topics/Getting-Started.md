# Getting Started

## Install
Install the cli and library by running the following commands in your terminal. 
```bash
go get github.com/kibu-sh/kibu
go install github.com/kibu-sh/kibu/cmd/kibue@main
go install github.com/google/wire/cmd/wire@latest
kibue init # does not work right now
# should copy .kibu/ and cue.mod to your project
# wire generate is required to produce the first version with a generate directive
```

### kibue.gen.go
```go
package discern

//go:generate go run -mod=readonly github.com/kibu-sh/kibu/cmd/kibue build ./src/backend/systems/...
```

## Recommended Structure
```
mkdir -p .kibu/store
mkdir -p src/backend/cmd/server
mkdir -p src/backend/systems/health
mkdir -p src/backend/systems/health/healthspec
touch docker-compose.yml
touch .kibu/workspace.cue
touch src/backend/cmd/server/server.go
touch src/backend/cmd/server/foreman.go
touch src/backend/cmd/server/wire.go
touch src/backend/cmd/server/wire_set.go
touch src/backend/systems/health/health.go
touch src/backend/systems/health/healthspec/health_dto.go
```

Once the wire.go and wire_set.go files have been created cd into the server directory and run wire once.