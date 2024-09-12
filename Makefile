go.generate:
	@echo "Generating mocks..."
	@go generate ./...
.PHONY: go.generate


go.build:
	@echo "Building..."
	@go build ./...
.PHONY: go.build

install:
	@go install ./cmd/kibue
.PHONY: install