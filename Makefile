go.generate:
	@echo "Generating mocks..."
	@go generate ./...
.PHONY: go.generate


go.build:
	@echo "Building..."
	@go build ./...
.PHONY: go.build

go.install:
	go install ./cmd/devx/devx.go
.PHONY: go.install