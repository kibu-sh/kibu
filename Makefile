go.generate:
	@echo "Generating mocks..."
	@go generate ./...
.PHONY: go.generate


go.build:
	@echo "Building..."
	@go build ./...
.PHONY: go.build

generate.cmd:
	go generate ./cmd/...
.PHONY: generate.cmd

install: generate.cmd
	@go install ./cmd/kibu
.PHONY: install